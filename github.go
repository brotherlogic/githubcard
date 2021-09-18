package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/brotherlogic/goserver"
	"github.com/brotherlogic/goserver/utils"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/brotherlogic/cardserver/card"
	pbgh "github.com/brotherlogic/githubcard/proto"
	pbgs "github.com/brotherlogic/goserver/proto"
	kmpb "github.com/brotherlogic/keymapper/proto"
	ppb "github.com/brotherlogic/proxy/proto"
)

const (
	// CONFIG where we store la config
	CONFIG = "/github.com/brotherlogic/githubcard/config"
)

type silence struct {
	silence string
	origin  string
}

// GithubBridge the bridge to the github API
type GithubBridge struct {
	*goserver.GoServer
	accessCode     string
	serving        bool
	getter         httpGetter
	attempts       int
	fails          int
	added          map[string]time.Time
	addedMutex     *sync.Mutex
	issues         []*pbgh.Issue
	silencedAlerts int
	silences       []string
	blankAlerts    int
	config         *pbgh.Config
	gets           int64
	posts          int64
	webhookcount   int64
	issueCount     int64
	addedCount     map[string]int64
	lastIssue      time.Time
	issueLock      *sync.Mutex
	githubsecret   string
	external       string
}

type httpGetter interface {
	Post(url string, data string) (*http.Response, error)
	Get(url string) (*http.Response, error)
	Delete(url string) (*http.Response, error)
	Patch(url string, data string) (*http.Response, error)
	Put(url string, data string) (*http.Response, error)
}

type prodHTTPGetter struct {
	accessToken string
}

func (h prodHTTPGetter) getClient() *http.Client {
	//ctx, cancel := utils.ManualContext("getclient", "getclient", time.Minute)
	//defer cancel()
	//return oauth2.NewClient(ctx, oauth2.StaticTokenSource(
	//	&oauth2.Token{AccessToken: h.accessToken}))
	return &http.Client{}
}

func (h prodHTTPGetter) prepRequest(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth("brotherlogic", h.accessToken)
}

func (h prodHTTPGetter) Post(url string, data string) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(data)))
	if err != nil {
		return nil, err
	}
	h.prepRequest(req)

	return h.getClient().Do(req)
}

func (h prodHTTPGetter) Patch(url string, data string) (*http.Response, error) {
	req, _ := http.NewRequest("PATCH", url, bytes.NewBuffer([]byte(data)))

	h.prepRequest(req)

	resp, err := h.getClient().Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode == 404 {
			return nil, status.Errorf(codes.NotFound, "patch returned %v -> %v", resp.StatusCode, string(body))
		}
		return nil, fmt.Errorf("patch returned %v -> %v", resp.StatusCode, string(body))
	}

	return resp, err
}

func (h prodHTTPGetter) Put(url string, data string) (*http.Response, error) {
	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer([]byte(data)))

	h.prepRequest(req)

	return h.getClient().Do(req)
}

func (h prodHTTPGetter) Delete(url string) (*http.Response, error) {
	req, _ := http.NewRequest("DELETE", url, bytes.NewBuffer([]byte{}))

	h.prepRequest(req)

	return h.getClient().Do(req)
}

func (h prodHTTPGetter) Get(url string) (*http.Response, error) {
	req, _ := http.NewRequest("GET", url, nil)

	h.prepRequest(req)

	return h.getClient().Do(req)
}

//Init a record getter
func Init() *GithubBridge {
	s := &GithubBridge{
		GoServer:   &goserver.GoServer{},
		serving:    true,
		getter:     prodHTTPGetter{},
		attempts:   0,
		fails:      0,
		added:      make(map[string]time.Time),
		addedMutex: &sync.Mutex{},
		config:     &pbgh.Config{},
		addedCount: make(map[string]int64),
		issueLock:  &sync.Mutex{},
	}
	return s
}

// DoRegister does RPC registration
func (b *GithubBridge) DoRegister(server *grpc.Server) {
	pbgh.RegisterGithubServer(server, b)
}

// ReportHealth alerts if we're not healthy
func (b GithubBridge) ReportHealth() bool {
	return true
}

func (b *GithubBridge) saveIssues(ctx context.Context, config *pbgh.Config) error {
	if config.ExternalIP == "" {
		log.Fatalf("Trying to save without IP: %v", config)
	}
	return b.KSclient.Save(ctx, CONFIG, config)
}

func (b *GithubBridge) readIssues(ctx context.Context) (*pbgh.Config, error) {
	config := &pbgh.Config{}
	data, _, err := b.KSclient.Read(ctx, CONFIG, config)
	if err != nil {
		return nil, err
	}
	config = data.(*pbgh.Config)

	if len(config.JobsOfInterest) == 0 {
		config.JobsOfInterest = append(config.JobsOfInterest, "githubreceiver")
	}

	if config.ExternalIP == "" {
		b.RaiseIssue("Missing ext", fmt.Sprintf("The external IP is missing?"))
	}

	if config.GetTitleToIssue() == nil {
		config.TitleToIssue = make(map[string]string)
	}

	return config, nil
}

// Shutdown shuts down the server
func (b *GithubBridge) Shutdown(ctx context.Context) error {
	return nil
}

// Mote promotes this server
func (b *GithubBridge) Mote(ctx context.Context, master bool) error {
	return nil
}

// GetState gets the state of the server
func (b *GithubBridge) GetState() []*pbgs.State {
	return []*pbgs.State{}
}

const (
	wait = 5 * time.Minute // Wait five minute between runs
)

func (b *GithubBridge) postURL(url string, data string) (*http.Response, error) {
	b.posts++
	return b.getter.Post(url, data)
}

func (b *GithubBridge) patchURL(url string, data string) (*http.Response, error) {
	b.posts++
	return b.getter.Patch(url, data)
}

func (b *GithubBridge) putURL(url string, data string) (*http.Response, error) {
	b.posts++
	return b.getter.Put(url, data)
}

func (b *GithubBridge) deleteURL(url string) (*http.Response, error) {
	b.posts++
	return b.getter.Delete(url)
}

func (b *GithubBridge) visitURL(url string) (string, bool, error) {
	b.gets++

	resp, err := b.getter.Get(url)
	if err != nil {
		return "", false, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", false, err
	}

	if resp.StatusCode != 200 && resp.StatusCode != 0 {
		b.Log(fmt.Sprintf("Error in visit (%v): %v", resp.StatusCode, string(body)))
		return string(body), false, fmt.Errorf("Non 200 return (%v) -> %v", resp.StatusCode, string(body))
	}

	return string(body), len(resp.Header["Link"]) >= 1, nil
}

// Project is a project in the github world
type Project struct {
	Name string
}

// Webhook struct describing a simple webhook
type Webhook struct {
	ID        int      `json:"id"`
	Name      string   `json:"name"`
	Active    bool     `json:"active"`
	Events    []string `json:"events"`
	AddEvents []string `json:"add_events"`
	Config    Config   `json:"config"`
}

// WebhookAdd struct describing a simple webhook
type WebhookAdd struct {
	AddEvents []string `json:"add_events"`
	Config    Config   `json:"config"`
}

// Config struct for webhook
type Config struct {
	URL         string `json:"url"`
	ContentType string `json:"content_type"`
	InsecureSSL string `json:"insecure_ssl"`
	Secret      string `json:"secret"`
}

func (b *GithubBridge) getWebHooks(ctx context.Context, repo string) ([]*Webhook, error) {
	urlv := fmt.Sprintf("https://api.github.com/repos/brotherlogic/%v/hooks", repo)
	body, _, err := b.visitURL(urlv)

	if err != nil {
		return []*Webhook{}, err
	}

	var data []*Webhook
	err = json.Unmarshal([]byte(body), &data)
	if err != nil {
		return []*Webhook{}, err
	}

	result := make([]*Webhook, 0)
	for _, d := range data {
		if d.Name == "web" {
			result = append(result, d)
		}
	}

	return result, nil
}

func (b *GithubBridge) updateWebHook(ctx context.Context, repo string, hook *Webhook) error {
	urlv := fmt.Sprintf("https://api.github.com/repos/brotherlogic/%v/hooks/%v", repo, hook.ID)

	nhook := &WebhookAdd{AddEvents: hook.Events, Config: Config{ContentType: hook.Config.ContentType, Secret: hook.Config.Secret, URL: hook.Config.URL}}
	bytes, err := json.Marshal(nhook)
	if err != nil {
		return err
	}

	resp, err := b.patchURL(urlv, string(bytes))
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	b.Log(fmt.Sprintf("RESULT: %v", string(data)))

	return err
}

func (b *GithubBridge) addWebHook(ctx context.Context, repo string, hook Webhook) error {
	urlv := fmt.Sprintf("https://api.github.com/repos/brotherlogic/%v/hooks", repo)

	bytes, err := json.Marshal(hook)
	if err != nil {
		return err
	}

	resp, err := b.postURL(urlv, string(bytes))
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)

	b.Log(fmt.Sprintf("READ[%v]: %v", resp.StatusCode, string(data)))

	return err
}

func (b *GithubBridge) issueExists(title string) (*pbgh.Issue, error) {
	urlv := "https://api.github.com/user/issues"
	body, _, err := b.visitURL(urlv)

	if err != nil {
		return nil, err
	}

	var data []interface{}
	err = json.Unmarshal([]byte(body), &data)
	if err != nil {
		return nil, err
	}

	b.issueCount = int64(len(data))
	var retIssue *pbgh.Issue
	seenUrls := make(map[string]bool)
	for _, d := range data {
		dp := d.(map[string]interface{})
		if dp["title"].(string) == title {
			retIssue = &pbgh.Issue{Title: title}
		}

		found := false
		for _, issue := range b.config.Issues {
			if dp["url"].(string) == issue.Url {
				t, _ := time.Parse("2006-01-02T15:04:05Z", dp["created_at"].(string))
				issue.DateAdded = t.Unix()
				issue.Title = dp["title"].(string)
				found = true
			}
		}

		if !found {
			val, _ := strconv.Atoi(dp["created_at"].(string))
			b.config.Issues = append(b.config.Issues, &pbgh.Issue{Title: dp["title"].(string), Url: dp["url"].(string), DateAdded: int64(val)})
		}

		seenUrls[dp["url"].(string)] = true
	}

	for i, issue := range b.config.Issues {
		if !seenUrls[issue.Url] {
			b.config.Issues = append(b.config.Issues[:i], b.config.Issues[i+1:]...)
			return retIssue, nil
		}
	}

	return retIssue, nil
}

// AmRequest milestone request
type AmRequest struct {
	Title       string `json:"title"`
	State       string `json:"state"`
	Description string `json:"description"`
}

// AmResponse milestone add response
type AmResponse struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
}

func (b *GithubBridge) getMilestoneLocal(ctx context.Context, repo, title string) (int, error) {
	urlv := fmt.Sprintf("https://api.github.com/repos/brotherlogic/%v/milestones", repo)

	resp, err := b.getter.Get(urlv)

	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()
	rb, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 && resp.StatusCode != 0 {
		return -1, fmt.Errorf("Unable to get milestones: %v->%v", resp.StatusCode, string(rb))
	}

	var amresponse []AmResponse
	err = json.Unmarshal([]byte(rb), &amresponse)
	if err != nil {
		return -1, err
	}

	for _, amresp := range amresponse {
		if amresp.Title == title {
			return amresp.Number, nil
		}
	}

	return -1, fmt.Errorf("Cannot locate milestone (%v)", title)
}

func (b *GithubBridge) createMilestoneLocal(ctx context.Context, repo, title, state, description string) (int, error) {
	urlv := fmt.Sprintf("https://api.github.com/repos/brotherlogic/%v/milestones", repo)

	payload := &AmRequest{Title: title, State: state, Description: description}
	bytes, err := json.Marshal(payload)
	if err != nil {
		return -1, err
	}

	resp, err := b.postURL(urlv, string(bytes))
	if err != nil {
		return -1, err
	}

	// Possible double add
	if resp.StatusCode == 422 {
		return b.getMilestoneLocal(ctx, repo, title)
	}

	if resp.StatusCode != 200 && resp.StatusCode != 201 && resp.StatusCode != 0 {
		defer resp.Body.Close()
		rb, _ := ioutil.ReadAll(resp.Body)

		return -1, fmt.Errorf("Unable to add milestone: %v -> %v", resp.StatusCode, string(rb))
	}

	defer resp.Body.Close()
	rb, _ := ioutil.ReadAll(resp.Body)

	var amresponse *AmResponse
	err = json.Unmarshal([]byte(rb), &amresponse)
	if err != nil {
		return -1, err
	}

	return amresponse.Number, err
}

// UmRequest milestone request
type UmRequest struct {
	State string `json:"state"`
}

func (b *GithubBridge) updateMilestoneLocal(ctx context.Context, repo string, number int32, state string) error {
	urlv := fmt.Sprintf("https://api.github.com/repos/brotherlogic/%v/milestones/%v", repo, number)

	payload := &UmRequest{State: state}
	bytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := b.postURL(urlv, string(bytes))
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 && resp.StatusCode != 201 && resp.StatusCode != 0 {
		return fmt.Errorf("UNable to update milestone: %v", resp.StatusCode)
	}

	return err
}

// PRequest pull request
type PRequest struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Head  string `json:"head"`
	Base  string `json:"base"`
}

type Labels struct {
	Labels []string `json:"labels"`
}

type PResponse struct {
	number int32 `json:number`
}

func (b *GithubBridge) createPullRequestLocal(ctx context.Context, job, branch, title string) (int32, error) {
	urlv := fmt.Sprintf("https://api.github.com/repos/brotherlogic/%v/pulls", job)

	payload := &PRequest{Title: title, Head: branch, Base: "master", Body: "Auto created pull request"}
	bytes, err := json.Marshal(payload)
	if err != nil {
		return -1, err
	}

	resp, err := b.postURL(urlv, string(bytes))
	if err != nil {
		return -1, err
	}

	if resp.StatusCode != 200 && resp.StatusCode != 201 && resp.StatusCode != 0 {
		return -1, fmt.Errorf("UNable to build pull request: %v", resp.StatusCode)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return -1, err
	}

	presp := &PResponse{}
	err = json.Unmarshal(body, presp)

	return presp.number, err
}

func (b *GithubBridge) addLabel(ctx context.Context, job, branch, title string, number int32, label string) error {
	urlv := fmt.Sprintf("https://api.github.com/repos/brotherlogic/%v/pulls/%v/labels", job, number)

	payload := &Labels{Labels: []string{"automerge"}}
	bytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := b.postURL(urlv, string(bytes))
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 && resp.StatusCode != 201 && resp.StatusCode != 0 {
		return fmt.Errorf("Unable to set label: %v", resp.StatusCode)
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	b.Log(fmt.Sprintf("Adding LABEL: %v -> %v", urlv, string(body)))

	return nil
}

type commit struct {
	Sha string `json:"sha"`
}

type prr struct {
	State string `json:"state"`
}

func (b *GithubBridge) getPullRequestLocal(ctx context.Context, job string, pullNumber int32) (*pbgh.PullResponse, error) {
	urlv := fmt.Sprintf("https://api.github.com/repos/brotherlogic/%v/pulls/%v/commits", job, pullNumber)
	body, _, err := b.visitURL(urlv)
	if err != nil {
		return nil, err
	}

	var data []*commit
	err = json.Unmarshal([]byte(body), &data)
	if err != nil {
		return nil, err
	}

	urlv = fmt.Sprintf("https://api.github.com/repos/brotherlogic/%v/pulls/%v", job, pullNumber)
	body, _, err = b.visitURL(urlv)
	if err != nil {
		return nil, err
	}

	var prdata *prr
	err = json.Unmarshal([]byte(body), &prdata)
	if err != nil {
		return nil, err
	}

	return &pbgh.PullResponse{NumberOfCommits: int32(len(data)), IsOpen: prdata.State == "open"}, nil
}

type closePayload struct {
	Sha string `json:"sha"`
}

func (b *GithubBridge) closePullRequestLocal(ctx context.Context, job string, pullNumber int32, sha string) (*pbgh.CloseResponse, error) {
	urlv := fmt.Sprintf("https://api.github.com/repos/brotherlogic/%v/pulls/%v/merge", job, pullNumber)

	payload := closePayload{Sha: sha}
	bytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	resp, err := b.putURL(urlv, string(bytes))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return nil, fmt.Errorf("Error closing pull request: %v", resp.StatusCode)
	}

	return &pbgh.CloseResponse{}, nil
}

func (b *GithubBridge) deleteBranchLocal(ctx context.Context, job string, branchName string) error {
	urlv := fmt.Sprintf("https://api.github.com/repos/brotherlogic/%v/git/refs/heads/%v", job, branchName)

	_, err := b.deleteURL(urlv)
	return err
}

// Payload for sending to github
type Payload struct {
	Title    string `json:"title"`
	Body     string `json:"body"`
	Assignee string `json:"assignee"`
}

// PayloadWithMilestone same as above but with a milestone field.
type PayloadWithMilestone struct {
	Title     string `json:"title"`
	Body      string `json:"body"`
	Assignee  string `json:"assignee"`
	Milestone int    `json:"milestone"`
}

type ClosePayload struct {
	State string `json:"state"`
}

func (b *GithubBridge) DeleteIssueLocal(ctx context.Context, owner string, issue *pbgh.Issue) error {
	payload := ClosePayload{State: "closed"}
	bytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	b.Log(fmt.Sprintf("Deleting issue %v/%v", issue.GetService(), issue.GetNumber()))
	_, err = b.patchURL(fmt.Sprintf("https://api.github.com/repos/%v/%v/issues/%v", owner, issue.GetService(), issue.GetNumber()), string(bytes))
	return err
}

// AddIssueLocal adds an issue
func (b *GithubBridge) AddIssueLocal(owner, repo, title, body string, milestone int) ([]byte, error) {
	b.attempts++
	issue, err := b.issueExists(title)
	if err != nil {
		return nil, err
	}
	if issue != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "Issue already exists")
	}

	b.Log(fmt.Sprintf("Adding Issue: %v", title))

	payload := Payload{Title: title, Body: body, Assignee: owner}
	bytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	if milestone > 0 {
		payload := PayloadWithMilestone{Title: title, Body: body, Assignee: owner, Milestone: milestone}
		bytes, err = json.Marshal(payload)
		if err != nil {
			return nil, err
		}

	}

	urlv := "https://api.github.com/repos/" + owner + "/" + repo + "/issues"
	resp, err := b.postURL(urlv, string(bytes))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	rb, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 && resp.StatusCode != 201 && resp.StatusCode != 0 {
		return rb, fmt.Errorf("POST error: %v -> %v", resp.StatusCode, string(rb))
	}

	return rb, nil
}

func hash(s string) int32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return int32(h.Sum32())
}

// GetIssueLocal Gets github issues for a given project
func (b *GithubBridge) GetIssueLocal(ctx context.Context, owner string, project string, number int) (*pbgh.Issue, error) {
	urlv := "https://api.github.com/repos/" + owner + "/" + project + "/issues/" + strconv.Itoa(number)
	body, _, err := b.visitURL(urlv)

	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	err = json.Unmarshal([]byte(body), &data)
	if err != nil {
		return nil, err
	}

	if _, ok := data["title"]; !ok {
		b.RaiseIssue("Bad parse", fmt.Sprintf("Bad parse %v", string(body)))
	}

	issue := &pbgh.Issue{Number: int32(number), Service: project, Title: data["title"].(string), Body: data["body"].(string)}
	if data["state"].(string) == "open" {
		issue.State = pbgh.Issue_OPEN
	} else {
		issue.State = pbgh.Issue_CLOSED
	}

	return issue, nil
}

// GetIssues Gets github issues for a given project
func (b *GithubBridge) GetIssues() pb.CardList {
	cardlist := pb.CardList{}
	urlv := "https://api.github.com/issues?state=open&filter=all"
	body, _, err := b.visitURL(urlv)

	if err != nil {
		return cardlist
	}

	var data []interface{}
	err = json.Unmarshal([]byte(body), &data)
	if err != nil {
		return cardlist
	}

	for _, issue := range data {
		issueMap := issue.(map[string]interface{})

		if _, ok := issueMap["pull_request"]; !ok {
			issueSource := issueMap["url"].(string)
			issueTitle := issueMap["title"].(string)
			issueText := issueMap["body"].(string)

			date, _ := time.Parse("2006-01-02T15:04:05Z", issueMap["created_at"].(string))

			card := &pb.Card{}
			card.Text = issueTitle + "\n" + issueText + "\n\n" + issueSource
			card.Hash = "githubissue-" + issueSource
			card.Channel = pb.Card_ISSUES
			card.Priority = int32(time.Now().Sub(date).Seconds())
			cardlist.Cards = append(cardlist.Cards, card)
		}
	}

	return cardlist
}

func (b *GithubBridge) cleanAdded(ctx context.Context) error {
	b.addedMutex.Lock()
	defer b.addedMutex.Unlock()
	for k, t := range b.added {
		if time.Now().Sub(t) > time.Minute {
			delete(b.added, k)
		}
	}

	return nil
}

func (b *GithubBridge) rebuild(ctx context.Context) error {
	_, err := b.issueExists("Clear Email")
	return err
}

func main() {
	var quiet = flag.Bool("quiet", true, "Show all output")
	var token = flag.String("token", "", "The token to use to auth")
	var external = flag.String("external", "", "External IP")
	var verify = flag.String("verify", "", "Token to use to verify")
	flag.Parse()

	b := Init()

	//Turn off logging
	if *quiet {
		log.SetFlags(0)
		log.SetOutput(ioutil.Discard)
	}

	b.PrepServer()
	b.Register = b

	if len(*verify) > 0 {
		b.getter = &prodHTTPGetter{accessToken: *verify}
		v, err := b.issueExists("Want Processing Needed!")
		fmt.Printf("%v and %v", err, v)
		return
	}

	err := b.RegisterServerV2("githubcard", false, true)
	if err != nil {
		return
	}

	ctx, cancel := utils.ManualContext("ghc", time.Minute)
	conn, err := b.FDialServer(ctx, "keymapper")
	if err != nil {
		if status.Convert(err).Code() == codes.Unknown {
			log.Fatalf("Cannot reach keymapper: %v", err)
		}
		return
	}
	client := kmpb.NewKeymapperServiceClient(conn)
	resp, err := client.Get(ctx, &kmpb.GetRequest{Key: "github_external"})
	if err != nil {
		if status.Convert(err).Code() == codes.Unknown || status.Convert(err).Code() == codes.InvalidArgument {
			log.Fatalf("Cannot read external: %v", err)
		}
		return
	}
	b.external = resp.GetKey().GetValue()
	cancel()

	if len(*token) > 0 {
		//b.Save(context.Bakground(), "/github.com/brotherlogic/githubcard/token", &pbgh.Token{Token: *token})
	} else if len(*external) > 0 {
		ctx, cancel := utils.ManualContext("githubc", time.Minute)
		defer cancel()
		config := &pbgh.Config{}
		data, _, err := b.KSclient.Read(ctx, CONFIG, config)
		if err != nil {
			log.Fatalf("%v", err)
		}
		tconfig := data.(*pbgh.Config)
		tconfig.ExternalIP = *external
		fmt.Printf("SAVED = %v\n", b.KSclient.Save(ctx, CONFIG, tconfig))
	} else {
		ctx, cancel := utils.ManualContext("githubs", time.Minute)
		m, _, err := b.Read(ctx, "/github.com/brotherlogic/githubcard/token", &pbgh.Token{})
		if err != nil {
			log.Fatalf("Error reading token: %v", err)
		}
		cancel()
		if len(m.(*pbgh.Token).GetToken()) == 0 {
			log.Fatalf("Error reading token: %v", m)
		}
		b.accessCode = m.(*pbgh.Token).GetToken()
		b.getter = &prodHTTPGetter{accessToken: b.accessCode}

		ctx, cancel = utils.ManualContext("githubs", time.Minute)
		m, _, err = b.Read(ctx, "/github.com/brotherlogic/github/secret", &ppb.GithubKey{})
		if err != nil {
			log.Fatalf("Error reading token: %v", err)
		}
		cancel()
		if len(m.(*ppb.GithubKey).GetKey()) == 0 {
			log.Fatalf("Error reading key: %v", m)
		}
		b.githubsecret = m.(*ppb.GithubKey).GetKey()

		b.Serve()
	}
}
