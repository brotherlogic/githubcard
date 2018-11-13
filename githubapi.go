package main

import (
	"encoding/json"
	"fmt"
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
	//Don't double add issues
	g.addedMutex.Lock()
	defer g.addedMutex.Unlock()
	if v, ok := g.added[in.GetTitle()]; ok {
		if !in.Sticky {
			return nil, fmt.Errorf("Unable to add this issue - recently added (%v)", v)
		}
		g.issues = append(g.issues, in)
		g.saveIssues(ctx)
		return in, nil
	}

	g.added[in.GetTitle()] = time.Now()
	b, err := g.AddIssueLocal("brotherlogic", in.GetService(), in.GetTitle(), in.GetBody())
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
