package main

import (
	"fmt"

	"golang.org/x/net/context"
)

func (g *GithubBridge) procSticky(ctx context.Context) error {
	for in, i := range g.issues {
		_, err := g.AddIssueLocal("brotherlogic", i.GetService(), i.GetTitle(), i.GetBody())
		if err == nil {
			g.issues = append(g.issues[:in], g.issues[in+1:]...)
			g.saveIssues(ctx)
			return nil
		}
	}

	return nil
}

func (g *GithubBridge) validateJobs(ctx context.Context) error {
	for _, j := range g.config.JobsOfInterest {
		g.Log(fmt.Sprintf("Checking %v", j))
	}

	return nil
}
