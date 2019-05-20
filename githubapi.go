package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"golang.org/x/net/context"

	pb "github.com/brotherlogic/githubcard/proto"
)

type addResponse struct {
	Number  int32
	Message string
}

//AddIssue adds an issue to github
func (g *GithubBridge) AddIssue(ctx context.Context, in *pb.Issue) (*pb.Issue, error) {
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
	if v, ok := g.added[in.GetTitle()]; ok {
		g.addedMutex.Unlock()
		if !in.Sticky {
			return nil, fmt.Errorf("Unable to add this issue - recently added (%v)", v)
		}
		g.issues = append(g.issues, in)
		g.saveIssues(ctx)
		return in, nil
	}
	g.added[in.GetTitle()] = time.Now()
	g.addedMutex.Unlock()

	b, err := g.AddIssueLocal("brotherlogic", in.GetService(), in.GetTitle(), in.GetBody())
	if err != nil {
		if in.Sticky {
			g.issues = append(g.issues, in)
			return in, nil
		}
		return nil, err
	}
	r := &addResponse{}
	log.Printf("UNMARSHAL: %v", string(b))
	err2 := json.Unmarshal(b, &r)
	if err2 != nil {
		return nil, err2
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

// Silence an issue
func (g *GithubBridge) Silence(ctx context.Context, in *pb.SilenceRequest) (*pb.SilenceResponse, error) {

	if in.Origin == "" {
		return nil, fmt.Errorf("Silence needs an origin")
	}

	currSilence := -1
	for i, sil := range g.config.Silences {
		if sil.Silence == in.Silence && sil.Origin == in.Origin {
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

	return nil, fmt.Errorf("Unable to apply request %v", in)
}
