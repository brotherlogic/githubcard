package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/brotherlogic/goserver"
	"github.com/brotherlogic/keystore/client"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/brotherlogic/cardserver/card"
	pbgh "github.com/brotherlogic/githubcard/proto"
	pbgs "github.com/brotherlogic/goserver/proto"
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
}

type httpGetter interface {
	Post(url string, data string) (*http.Response, error)
	Get(url string) (*http.Response, error)
	Patch(url string, data string) (*http.Response, error)
	Put(url string, data string) (*http.Response, error)
}

type prodHTTPGetter struct{}

func (httpGetter prodHTTPGetter) Post(url string, data string) (*http.Response, error) {
	return http.Post(url, "application/json", bytes.NewBuffer([]byte(data)))
}

func (httpGetter prodHTTPGetter) Patch(url string, data string) (*http.Response, error) {
	req, _ := http.NewRequest("PATCH", url, bytes.NewBuffer([]byte(data)))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	return client.Do(req)
}

func (httpGetter prodHTTPGetter) Put(url string, data string) (*http.Response, error) {
	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer([]byte(data)))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	return client.Do(req)
}

func (httpGetter prodHTTPGetter) Get(url string) (*http.Response, error) {
	return http.Get(url)
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

func (b *GithubBridge) saveIssues(ctx context.Context) {
	b.KSclient.Save(ctx, CONFIG, b.config)
}

func (b *GithubBridge) readIssues(ctx context.Context) error {
	config := &pbgh.Config{}
	data, _, err := b.KSclient.Read(ctx, CONFIG, config)
	if err != nil {
		return err
	}
	b.config = data.(*pbgh.Config)

	if len(b.config.JobsOfInterest) == 0 {
		b.config.JobsOfInterest = append(b.config.JobsOfInterest, "githubreceiver")
	}

	return nil
}

// Shutdown shuts down the server
func (b *GithubBridge) Shutdown(ctx context.Context) error {
	b.saveIssues(ctx)
	return nil
}

// Mote promotes this server
func (b *GithubBridge) Mote(ctx context.Context, master bool) error {
	if master {
		m, _, err := b.Read(ctx, "/github.com/brotherlogic/githubcard/token", &pbgh.Token{})
		if err != nil {
			return err
		}
		if len(m.(*pbgh.Token).GetToken()) == 0 {
			return fmt.Errorf("Error reading token: %v", m)
		}
		b.accessCode = m.(*pbgh.Token).GetToken()

		b.Log(fmt.Sprintf("READ: %v", b.accessCode))

		return b.readIssues(ctx)
	}
	return nil
}

// GetState gets the state of the server
func (b *GithubBridge) GetState() []*pbgs.State {
	b.addedMutex.Lock()
	defer b.addedMutex.Unlock()

	bestIssue := ""
	bestTime := time.Now().Unix()

	for _, issue := range b.config.Issues {
		if issue.DateAdded < bestTime {
			bestIssue = issue.Url
			bestTime = issue.DateAdded
		}
	}

	return []*pbgs.State{
		&pbgs.State{Key: "issues", Value: b.issueCount},
		&pbgs.State{Key: "current_issue", Text: bestIssue},
		&pbgs.State{Key: "webhook_count", Value: b.webhookcount},
		&pbgs.State{Key: "external", Text: b.config.ExternalIP},
		&pbgs.State{Key: "gets", Value: b.gets},
		&pbgs.State{Key: "posts", Value: b.posts},
		&pbgs.State{Key: "jobs", Text: fmt.Sprintf("%v", b.config.JobsOfInterest)},
		&pbgs.State{Key: "attempts", Value: int64(b.attempts)},
		&pbgs.State{Key: "fails", Value: int64(b.fails)},
		&pbgs.State{Key: "added", Text: fmt.Sprintf("%v", b.added)},
		&pbgs.State{Key: "sticky", Value: int64(len(b.issues))},
		&pbgs.State{Key: "silenced_alerts", Value: int64(b.silencedAlerts)},
		&pbgs.State{Key: "blank_alerts", Value: int64(b.blankAlerts)},
		&pbgs.State{Key: "silences", Text: fmt.Sprintf("%v", b.config.Silences)},
	}
}

const (
	wait = 5 * time.Minute // Wait five minute between runs
)

func (b *GithubBridge) postURL(urlv string, data string) (*http.Response, error) {
	url := urlv
	if len(b.accessCode) > 0 && strings.Contains(urlv, "?") {
		url = url + "&access_token=" + b.accessCode
	} else {
		url = url + "?access_token=" + b.accessCode
	}

	b.posts++
	b.Log(fmt.Sprintf("POST %v [%v]", url, data))
	return b.getter.Post(url, data)
}

func (b *GithubBridge) patchURL(urlv string, data string) (*http.Response, error) {
	url := urlv
	if len(b.accessCode) > 0 && strings.Contains(urlv, "?") {
		url = url + "&access_token=" + b.accessCode
	} else {
		url = url + "?access_token=" + b.accessCode
	}

	b.posts++
	b.Log(fmt.Sprintf("PATCH %v [%v]", url, data))
	return b.getter.Patch(url, data)
}

func (b *GithubBridge) putURL(urlv string, data string) (*http.Response, error) {
	url := urlv
	if len(b.accessCode) > 0 && strings.Contains(urlv, "?") {
		url = url + "&access_token=" + b.accessCode
	} else {
		url = url + "?access_token=" + b.accessCode
	}

	b.posts++
	b.Log(fmt.Sprintf("PUT %v [%v]", url, data))
	return b.getter.Put(url, data)
}

func (b *GithubBridge) visitURL(urlv string) (string, error) {

	url := urlv
	if len(b.accessCode) > 0 && strings.Contains(urlv, "?") {
		url = url + "&access_token=" + b.accessCode
	} else {
		url = url + "?access_token=" + b.accessCode
	}

	b.Log(fmt.Sprintf("VISIT %v", url))
	b.gets++
	resp, err := b.getter.Get(url)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	return string(body), nil
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
}

// Config struct for webhook
type Config struct {
	URL         string `json:"url"`
	ContentType string `json:"content_type"`
	InsecureSSL string `json:"insecure_ssl"`
}

func (b *GithubBridge) getWebHooks(ctx context.Context, repo string) ([]*Webhook, error) {
	urlv := fmt.Sprintf("https://api.github.com/repos/brotherlogic/%v/hooks", repo)
	body, err := b.visitURL(urlv)

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

	nhook := &WebhookAdd{AddEvents: hook.Events}
	bytes, err := json.Marshal(nhook)
	if err != nil {
		return err
	}

	resp, err := b.patchURL(urlv, string(bytes))
	if err != nil {
		return err
	}

	b.Log(fmt.Sprintf("UPDATE_WEB_HOOK = %v", resp))

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	b.Log(fmt.Sprintf("RESPONSE = %v (%v)", string(body), err))

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

	b.Log(fmt.Sprintf("ADD_WEB_HOOK = %v", resp))

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	b.Log(fmt.Sprintf("RESPONSE = %v (%v)", string(body), err))

	return err
}

func (b *GithubBridge) issueExists(title string) (*pbgh.Issue, error) {
	urlv := "https://api.github.com/user/issues"
	body, err := b.visitURL(urlv)

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
		b.Log(fmt.Sprintf("URL = %v", dp["url"]))

		found := false
		for _, issue := range b.config.Issues {
			if dp["url"].(string) == issue.Url {
				t, _ := time.Parse("2006-01-02T15:04:05Z", dp["created_at"].(string))
				issue.DateAdded = t.Unix()
				found = true
			}
		}

		if !found {
			val, _ := strconv.Atoi(dp["created_at"].(string))
			b.config.Issues = append(b.config.Issues, &pbgh.Issue{Title: title, Url: dp["url"].(string), DateAdded: int64(val)})
		}

		seenUrls[dp["url"].(string)] = true
	}

	b.Log(fmt.Sprintf("Pre-Remove %v and %v", len(b.config.Issues), len(data)))
	for i, issue := range b.config.Issues {
		if !seenUrls[issue.Url] {
			b.config.Issues = append(b.config.Issues[:i], b.config.Issues[i+1:]...)
			b.Log(fmt.Sprintf("Removing: %v", issue))
			return retIssue, nil
		}
	}

	return retIssue, nil
}

// PRequest pull request
type PRequest struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Head  string `json:"head"`
	Base  string `json:"base"`
}

func (b *GithubBridge) createPullRequestLocal(ctx context.Context, job, branch, title string) error {
	urlv := fmt.Sprintf("https://api.github.com/repos/brotherlogic/%v/pulls", job)

	payload := &PRequest{Title: title, Head: branch, Base: "master", Body: "Auto created pull request"}
	bytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := b.postURL(urlv, string(bytes))
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	rb, _ := ioutil.ReadAll(resp.Body)

	b.Log(fmt.Sprintf("PULL_REQUEST %v", string(rb)))

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return fmt.Errorf("UNable to build pull request: %v", resp.StatusCode)
	}

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
	body, err := b.visitURL(urlv)
	if err != nil {
		return nil, err
	}

	var data []*commit
	err = json.Unmarshal([]byte(body), &data)
	if err != nil {
		return nil, err
	}

	urlv = fmt.Sprintf("https://api.github.com/repos/brotherlogic/%v/pulls/%v", job, pullNumber)
	body, err = b.visitURL(urlv)
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

	defer resp.Body.Close()
	rb, _ := ioutil.ReadAll(resp.Body)
	b.Log(fmt.Sprintf("CLOSE %v", string(rb)))

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return nil, fmt.Errorf("Error closing pull request: %v", resp.StatusCode)
	}

	return &pbgh.CloseResponse{}, nil
}

// Payload for sending to github
type Payload struct {
	Title    string `json:"title"`
	Body     string `json:"body"`
	Assignee string `json:"assignee"`
}

// AddIssueLocal adds an issue
func (b *GithubBridge) AddIssueLocal(owner, repo, title, body string) ([]byte, error) {
	b.attempts++
	issue, err := b.issueExists(title)
	if err != nil {
		return nil, err
	}
	if issue != nil {
		return nil, errors.New("Issue already exists")
	}

	payload := Payload{Title: title, Body: body, Assignee: owner}
	bytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	b.Log(fmt.Sprintf("%v -> %v", payload, string(bytes)))

	urlv := "https://api.github.com/repos/" + owner + "/" + repo + "/issues"
	resp, err := b.postURL(urlv, string(bytes))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	rb, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		b.fails++
		b.Log(fmt.Sprintf("%v returned from github: %v -> %v", resp.StatusCode, string(rb), string(bytes)))
	}

	return rb, nil
}

func hash(s string) int32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return int32(h.Sum32())
}

// GetIssueLocal Gets github issues for a given project
func (b *GithubBridge) GetIssueLocal(owner string, project string, number int) (*pbgh.Issue, error) {
	urlv := "https://api.github.com/repos/" + owner + "/" + project + "/issues/" + strconv.Itoa(number)
	body, err := b.visitURL(urlv)
	b.Log(fmt.Sprintf("RETURN %v", body))

	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	err = json.Unmarshal([]byte(body), &data)
	if err != nil {
		return nil, err
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
	body, err := b.visitURL(urlv)

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

func main() {
	var quiet = flag.Bool("quiet", true, "Show all output")
	var token = flag.String("token", "", "The token to use to auth")
	var external = flag.String("external", "", "External IP")
	flag.Parse()

	b := Init()
	b.GoServer.KSclient = *keystoreclient.GetClient(b.DialMaster)

	//Turn off logging
	if *quiet {
		log.SetFlags(0)
		log.SetOutput(ioutil.Discard)
	}

	b.PrepServer()
	b.Register = b

	b.RegisterServer("githubcard", false)

	if len(*token) > 0 {
		//b.Save(context.Bakground(), "/github.com/brotherlogic/githubcard/token", &pbgh.Token{Token: *token})
	} else if len(*external) > 0 {
		/*config := &pbgh.Config{}
		data, _, err := b.KSclient.Read(context.Bacground(), CONFIG, config)
		if err != nil {
			log.Fatalf("%v", err)
		}
		tconfig := data.(*pbgh.Config)
		tconfig.ExternalIP = *external
		b.KSclient.Save(context.Bacground(), CONFIG, tconfig) */
	} else {
		b.RegisterRepeatingTask(b.cleanAdded, "clean_added", time.Minute)
		b.RegisterRepeatingTask(b.procSticky, "proc_sticky", time.Minute*5)
		b.RegisterRepeatingTask(b.validateJobs, "validate_jobs", time.Minute*5)
		b.Serve()
	}
}
