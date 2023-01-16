package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/brotherlogic/githubcard/proto"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type addResponse struct {
	Number  int32
	Message string
}

//ClosePullRequest closes a pull request
func (g *GithubBridge) ClosePullRequest(ctx context.Context, in *pb.CloseRequest) (*pb.CloseResponse, error) {
	resp, err := g.closePullRequestLocal(ctx, in.Job, in.PullNumber, in.Sha)
	if err != nil {
		return resp, err
	}
	err = g.deleteBranchLocal(ctx, in.Job, in.BranchName)
	return resp, err
}

//GetPullRequest gets a pull request
func (g *GithubBridge) GetPullRequest(ctx context.Context, in *pb.PullRequest) (*pb.PullResponse, error) {
	return g.getPullRequestLocal(ctx, in.Job, in.PullNumber)
}

//CreatePullRequest creates a pull request
func (g *GithubBridge) CreatePullRequest(ctx context.Context, in *pb.PullRequest) (*pb.PullResponse, error) {
	_, err := g.createPullRequestLocal(ctx, in.Job, in.Branch, in.Title)
	if err != nil {
		return nil, err
	}

	return &pb.PullResponse{}, err
}

//AddMilestone adds a milestone
func (g *GithubBridge) AddMilestone(ctx context.Context, req *pb.AddMilestoneRequest) (*pb.AddMilestoneResponse, error) {
	num, err := g.createMilestoneLocal(ctx, req.Repo, req.Title, "open", req.Description)
	return &pb.AddMilestoneResponse{Number: int32(num)}, err
}

//UpdateMilestone updates a milestone
func (g *GithubBridge) UpdateMilestone(ctx context.Context, req *pb.UpdateMilestoneRequest) (*pb.UpdateMilestoneResponse, error) {
	err := g.updateMilestoneLocal(ctx, req.Repo, req.Number, req.State)
	return &pb.UpdateMilestoneResponse{}, err
}

//RegisterJob registers a job to be built
func (g *GithubBridge) RegisterJob(ctx context.Context, in *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	return &pb.RegisterResponse{}, g.validateJob(ctx, in.Job)
}

//DeleteIssue removes an issue
func (g *GithubBridge) DeleteIssue(ctx context.Context, in *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	config, err := g.readIssues(ctx)
	if err != nil {
		return nil, err
	}

	issue := in.GetIssue()
	g.issueLock.Lock()
	defer g.issueLock.Unlock()
	for i, is := range config.Issues {
		if is.Service == in.Issue.Service && is.Number == in.Issue.Number {
			config.Issues = append(config.Issues[:i], config.Issues[i+1:]...)
			issue = is

			err := g.saveIssues(ctx, config)
			if err != nil {
				return nil, err
			}

			break
		}
	}

	if issue.GetPrintId() == 0 {
		g.CtxLog(ctx, fmt.Sprintf("Issue has no print id: %v", issue))
	}

	// Fire and forget to subscribers
	for _, subscriber := range issue.GetSubscribers() {
		conn, err := g.FDialServer(ctx, subscriber)
		defer conn.Close()
		if err == nil {
			client := pb.NewGithubSubscriberClient(conn)
			client.ChangeUpdate(ctx, &pb.ChangeUpdateRequest{Issue: issue})
		}
	}

	return &pb.DeleteResponse{}, g.DeleteIssueLocal(ctx, "brotherlogic", issue)
}

var (
	issues = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "githubcard_issues",
		Help: "The number of issues added per binary",
	}, []string{"service"})
)

//AddIssue adds an issue to github
func (g *GithubBridge) AddIssue(ctx context.Context, in *pb.Issue) (*pb.Issue, error) {
	// Lock the whole add process
	key, err := g.RunLockingElection(ctx, "github-issue", "Holding for issue")
	if err != nil {
		return nil, err
	}
	defer g.ReleaseLockingElection(ctx, "github-issue", key)

	//Don't double add issues
	g.addedMutex.Lock()
	g.addedCount[in.GetTitle()]++
	if v, ok := g.added[in.GetTitle()]; ok {
		if time.Since(v) < time.Minute*10 {
			g.addedMutex.Unlock()
			return nil, status.Errorf(codes.ResourceExhausted, "Unable to add this issue (%v)- recently added (%v)", in.GetTitle(), v)
		}
	}
	g.added[in.GetTitle()] = time.Now()
	g.addedMutex.Unlock()

	config, err := g.readIssues(ctx)
	if err != nil {
		return nil, err
	}

	g.lastIssue = time.Now()

	// If this comes from the receiver - just add it
	if in.Origin == pb.Issue_FROM_RECEIVER {
		issue := in
		for _, issues := range config.Issues {
			if issues.GetNumber() == issue.GetNumber() && issues.GetService() == issue.GetService() {
				issue = issues
			}
		}

		g.CtxLog(ctx, fmt.Sprintf("Resolved to existing issue: %v -> %v", issue, in))

		if issue.GetPrintId() == 0 {
			config.Issues = append(config.Issues, in)
		}

		issue.DateAdded = time.Now().Unix()
		return in, g.saveIssues(ctx, config)
	}

	// Reject alerts with a blank body
	if len(in.GetBody()) == 0 {
		g.blankAlerts++
		return &pb.Issue{}, fmt.Errorf("this issue has no body")
	}

	//Reject silenced issues
	for _, silence := range config.Silences {
		if in.GetTitle() == silence.Silence {
			g.silencedAlerts++
			return &pb.Issue{}, fmt.Errorf("this issue is silenced")
		}
	}

	//Reject any issue we've seen before
	for title, issue := range config.GetTitleToIssue() {
		if in.GetTitle() == title {
			// Is this title still open
			elems := strings.Split(issue, "/")
			num, _ := strconv.Atoi(elems[1])
			i, err := g.GetIssueLocal(ctx, "brotherlogic", elems[0], num)
			if err != nil {
				return nil, err
			}

			if i.State == pb.Issue_OPEN {
				return nil, status.Errorf(codes.AlreadyExists, "We already have an issue with this title: %v", issue)
			}

			delete(config.TitleToIssue, title)
		}
	}

	b, pid, err := g.AddIssueLocal(ctx, "brotherlogic", in.GetService(), in.GetTitle(), in.GetBody(), int(in.GetMilestoneNumber()), in.GetPrintImmediately(), config)
	if err != nil {
		if in.Sticky {
			g.issues = append(g.issues, in)
			return in, nil
		}
		return nil, err
	}
	r := &addResponse{}
	in.PrintId = pid

	err2 := json.Unmarshal(b, &r)
	if err2 != nil {
		return nil, fmt.Errorf("error unmarshal: %v from %v", err2, string(b))
	}

	if r.Message == "Not Found" {
		g.AddIssue(ctx, &pb.Issue{Service: "githubcard", Title: "Add Failure", Body: fmt.Sprintf("Couldn't add issue for %v with title %v (%v)", in.Service, in.GetTitle(), in.GetBody())})
		return nil, fmt.Errorf("error adding issue for service %v", in.Service)
	}

	g.CtxLog(ctx, fmt.Sprintf("Adding Issue: %v -> %v/%v (%v)", in.GetTitle(), in.GetService(), r.Number, in))

	in.Number = r.Number

	config.TitleToIssue[in.GetTitle()] = fmt.Sprintf("%v/%v", in.GetService(), in.GetNumber())
	mapSize.Set(float64(len(config.GetTitleToIssue())))
	config.Issues = append(config.Issues, in)
	return in, g.saveIssues(ctx, config)
}

var (
	mapSize = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "githubcard_map_size",
		Help: "The size of the print queue",
	})
)

//Get gets an issue from github
func (g *GithubBridge) Get(ctx context.Context, in *pb.Issue) (*pb.Issue, error) {
	b, err := g.GetIssueLocal(ctx, "brotherlogic", in.GetService(), int(in.GetNumber()))
	return b, err
}

//GetAll gets all the issues currently open
func (g *GithubBridge) GetAll(ctx context.Context, in *pb.GetAllRequest) (*pb.GetAllResponse, error) {
	config, err := g.readIssues(ctx)
	if err != nil {
		return nil, err
	}

	resp := &pb.GetAllResponse{}

	for _, is := range config.Issues {
		allowed := true
		g.CtxLog(ctx, fmt.Sprintf("BUT HERE %v", is))
		for _, no := range in.GetAvoid() {
			if strings.Contains(is.GetUrl(), no) {
				allowed = false
			}
		}

		if allowed && is.State != pb.Issue_CLOSED {
			resp.Issues = append(resp.Issues, is)
		}
	}

	sort.SliceStable(resp.Issues, func(i, j int) bool {
		return resp.Issues[i].DateAdded < resp.Issues[j].DateAdded
	})

	if in.LatestOnly {
		return &pb.GetAllResponse{Issues: resp.Issues[0:]}, nil
	}

	return resp, nil
}

// Silence an issue
func (g *GithubBridge) Silence(ctx context.Context, in *pb.SilenceRequest) (*pb.SilenceResponse, error) {
	config, err := g.readIssues(ctx)
	if err != nil {
		return nil, err
	}

	if in.Origin == "" {
		return nil, fmt.Errorf("Silence needs an origin")
	}

	currSilence := -1
	for i, sil := range config.Silences {
		if sil.Origin == in.Origin {
			currSilence = i
		}
	}

	if in.State == pb.SilenceRequest_SILENCE && currSilence == -1 {
		config.Silences = append(config.Silences, &pb.Silence{Silence: in.Silence, Origin: in.Origin})
		g.saveIssues(ctx, config)
		return &pb.SilenceResponse{}, nil
	}

	if in.State == pb.SilenceRequest_UNSILENCE && currSilence >= 0 {
		config.Silences = append(config.Silences[:currSilence], config.Silences[currSilence+1:]...)
		g.saveIssues(ctx, config)
		return &pb.SilenceResponse{}, nil
	}

	return nil, fmt.Errorf("unable to apply request, silence was not found %v", in)
}

//Configure the system
func (g *GithubBridge) Configure(ctx context.Context, req *pb.ConfigureRequest) (*pb.ConfigureResponse, error) {
	config, err := g.readIssues(ctx)
	if err != nil {
		return nil, err
	}
	config.ExternalIP = req.GetExternalIp()
	return &pb.ConfigureResponse{}, g.saveIssues(ctx, config)
}
