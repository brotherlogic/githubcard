package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/brotherlogic/githubcard/proto"
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
	err := g.createPullRequestLocal(ctx, in.Job, in.Branch, in.Title)
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
	for _, j := range g.config.JobsOfInterest {
		if j == in.Job {
			return &pb.RegisterResponse{}, nil
		}
	}

	g.config.JobsOfInterest = append(g.config.JobsOfInterest, in.Job)
	g.saveIssues(ctx)

	return &pb.RegisterResponse{}, nil
}

//DeleteIssue removes an issue
func (g *GithubBridge) DeleteIssue(ctx context.Context, in *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	for i, is := range g.config.Issues {
		if is.Url == in.Issue.Url {
			g.config.Issues = append(g.config.Issues[:i], g.config.Issues[i+1:]...)
		}
	}

	return &pb.DeleteResponse{}, nil
}

//AddIssue adds an issue to github
func (g *GithubBridge) AddIssue(ctx context.Context, in *pb.Issue) (*pb.Issue, error) {
	g.lastIssue = time.Now()

	// If this comes from the receiver - just add it
	if in.Origin == pb.Issue_FROM_RECEIVER {
		in.DateAdded = time.Now().Unix()
		g.config.Issues = append(g.config.Issues, in)
		g.saveIssues(ctx)
		return in, nil
	}

	// Reject alerts with a blank body
	if len(in.GetBody()) == 0 {
		g.blankAlerts++
		return &pb.Issue{}, fmt.Errorf("This issue has no body")
	}

	//Reject silenced issues
	for _, silence := range g.config.Silences {
		if in.GetTitle() == silence.Silence {
			g.silencedAlerts++
			return &pb.Issue{}, fmt.Errorf("This issue is silenced")
		}
	}

	//Don't double add issues
	g.addedMutex.Lock()
	g.addedCount[in.GetTitle()]++
	if v, ok := g.added[in.GetTitle()]; ok {
		g.addedMutex.Unlock()
		if !in.Sticky {
			return nil, status.Errorf(codes.ResourceExhausted, "Unable to add this issue (%v)- recently added (%v)", in.GetTitle(), v)
		}
		g.issues = append(g.issues, in)
		g.saveIssues(ctx)
		return in, nil
	}
	g.added[in.GetTitle()] = time.Now()
	g.addedMutex.Unlock()

	b, err := g.AddIssueLocal("brotherlogic", in.GetService(), in.GetTitle(), in.GetBody(), int(in.GetMilestoneNumber()))
	if err != nil {
		if in.Sticky {
			g.issues = append(g.issues, in)
			return in, nil
		}
		return nil, err
	}
	r := &addResponse{}

	err2 := json.Unmarshal(b, &r)
	if err2 != nil {
		return nil, fmt.Errorf("Error unmarshal: %v from %v", err2, string(b))
	}

	if r.Message == "Not Found" {
		g.AddIssue(ctx, &pb.Issue{Service: "githubcard", Title: "Add Failure", Body: fmt.Sprintf("Couldn't add issue for %v with title %v (%v)", in.Service, in.GetTitle(), in.GetBody())})
		return nil, fmt.Errorf("Error adding issue for service %v", in.Service)
	}

	in.Number = r.Number
	return in, nil
}

//Get gets an issue from github
func (g *GithubBridge) Get(ctx context.Context, in *pb.Issue) (*pb.Issue, error) {
	b, err := g.GetIssueLocal("brotherlogic", in.GetService(), int(in.GetNumber()))
	return b, err
}

//GetAll gets all the issues currently open
func (g *GithubBridge) GetAll(ctx context.Context, in *pb.GetAllRequest) (*pb.GetAllResponse, error) {
	resp := &pb.GetAllResponse{}

	for _, is := range g.config.Issues {
		resp.Issues = append(resp.Issues, is)
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

	if in.Origin == "" {
		return nil, fmt.Errorf("Silence needs an origin")
	}

	currSilence := -1
	for i, sil := range g.config.Silences {
		if sil.Origin == in.Origin {
			currSilence = i
		}
	}

	if in.State == pb.SilenceRequest_SILENCE && currSilence == -1 {
		g.config.Silences = append(g.config.Silences, &pb.Silence{Silence: in.Silence, Origin: in.Origin})
		g.saveIssues(ctx)
		return &pb.SilenceResponse{}, nil
	}

	if in.State == pb.SilenceRequest_UNSILENCE && currSilence >= 0 {
		g.config.Silences = append(g.config.Silences[:currSilence], g.config.Silences[currSilence+1:]...)
		g.saveIssues(ctx)
		return &pb.SilenceResponse{}, nil
	}

	return nil, fmt.Errorf("Unable to apply request, silence was not found %v", in)
}
