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
	// KEY the issues
	KEY = "/github.com/brotherlogic/githubcard/issues"
)

// GithubBridge the bridge to the github API
type GithubBridge struct {
	*goserver.GoServer
	accessCode string
	serving    bool
	getter     httpGetter
	attempts   int
	fails      int
	added      map[string]time.Time
	addedMutex *sync.Mutex
	issues     []*pbgh.Issue
}

type httpGetter interface {
	Post(url string, data string) (*http.Response, error)
	Get(url string) (*http.Response, error)
}

type prodHTTPGetter struct{}

func (httpGetter prodHTTPGetter) Post(url string, data string) (*http.Response, error) {
	return http.Post(url, "application/json", bytes.NewBuffer([]byte(data)))
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
	}
	s.Register = s
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
	b.KSclient.Save(ctx, KEY, &pbgh.IssueList{Issues: b.issues})
}

func (b GithubBridge) readIssues(ctx context.Context) error {
	issues := &pbgh.IssueList{}
	data, _, err := b.KSclient.Read(ctx, KEY, issues)
	if err != nil {
		return err
	}
	b.issues = (data.(*pbgh.IssueList).Issues)
	return nil
}

// Mote promotes this server
func (b GithubBridge) Mote(ctx context.Context, master bool) error {
	if master {
		return b.readIssues(ctx)
	}
	return nil
}

// GetState gets the state of the server
func (b GithubBridge) GetState() []*pbgs.State {
	b.addedMutex.Lock()
	defer b.addedMutex.Unlock()
	return []*pbgs.State{
		&pbgs.State{Key: "attempts", Value: int64(b.attempts)},
		&pbgs.State{Key: "fails", Value: int64(b.fails)},
		&pbgs.State{Key: "added", Text: fmt.Sprintf("%v", b.added)},
		&pbgs.State{Key: "sticky", Value: int64(len(b.issues))},
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

	return b.getter.Post(url, data)
}

func (b *GithubBridge) visitURL(urlv string) (string, error) {

	url := urlv
	if len(b.accessCode) > 0 && strings.Contains(urlv, "?") {
		url = url + "&access_token=" + b.accessCode
	} else {
		url = url + "?access_token=" + b.accessCode
	}

	b.Log(fmt.Sprintf("VISIT %v", url))
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

func (b *GithubBridge) issueExists(title string) (*pbgh.Issue, error) {
	urlv := "https://api.github.com/user/issues"
	body, err := b.visitURL(urlv)

	b.Log(fmt.Sprintf("RESULT = %v", body))

	if err != nil {
		return nil, err
	}

	var data []interface{}
	err = json.Unmarshal([]byte(body), &data)
	if err != nil {
		return nil, err
	}

	for _, d := range data {
		dp := d.(map[string]interface{})
		if dp["title"].(string) == title {
			return &pbgh.Issue{Title: title}, nil
		}
	}

	return nil, nil
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

	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	err = json.Unmarshal([]byte(body), &data)
	if err != nil {
		return nil, err
	}

	log.Printf("HERE: (%v %v) %v", data["state"].(string), data["state"].(string) == "open", data)
	issue := &pbgh.Issue{Number: int32(number), Service: project, Title: data["title"].(string), Body: data["body"].(string)}
	if data["state"].(string) == "open" {
		issue.State = pbgh.Issue_OPEN
	} else {
		issue.State = pbgh.Issue_CLOSED
	}

	log.Printf("ISSUE = %v", issue)
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

			date, err := time.Parse("2006-01-02T15:04:05Z", issueMap["created_at"].(string))

			if err != nil {
				log.Printf("Error reading dates: %v", err)
			}

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

func (b *GithubBridge) cleanAdded(ctx context.Context) {
	b.addedMutex.Lock()
	defer b.addedMutex.Unlock()
	for k, t := range b.added {
		if time.Now().Sub(t) > time.Minute {
			delete(b.added, k)
		}
	}
}

func main() {
	var quiet = flag.Bool("quiet", true, "Show all output")
	var token = flag.String("token", "", "The token to use to auth")
	flag.Parse()

	b := Init()
	b.GoServer.KSclient = *keystoreclient.GetClient(b.GetIP)

	//Turn off logging
	if *quiet {
		log.SetFlags(0)
		log.SetOutput(ioutil.Discard)
	}

	b.PrepServer()
	b.RegisterServer("githubcard", false)

	if len(*token) > 0 {
		b.Save(context.Background(), "/github.com/brotherlogic/githubcard/token", &pbgh.Token{Token: *token})
	} else {
		m, _, err := b.Read(context.Background(), "/github.com/brotherlogic/githubcard/token", &pbgh.Token{})
		if err != nil {
			log.Printf("Failed to read token: %v", err)
		} else {
			log.Printf("GOT TOKEN: %v", m)
			b.accessCode = m.(*pbgh.Token).GetToken()
			b.RegisterRepeatingTask(b.cleanAdded, "clean_added", time.Minute)
			b.RegisterRepeatingTask(b.procSticky, "proc_sticky", time.Minute*5)
			b.Serve()
		}
	}
}
