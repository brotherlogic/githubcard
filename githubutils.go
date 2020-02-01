package main

import (
	"fmt"
	"strings"
	"time"

	"golang.org/x/net/context"
)

func (g *GithubBridge) procSticky(ctx context.Context) error {
	for in, i := range g.issues {
		_, err := g.AddIssueLocal("brotherlogic", i.GetService(), i.GetTitle(), i.GetBody(), int(i.GetMilestoneNumber()))
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
		err := g.validateJob(ctx, j)

		if err != nil {
			g.Log(fmt.Sprintf("Error validating %v -> %v", j, err))
			return err
		}
	}

	g.Log(fmt.Sprintf("Validating %v jobs", len(g.config.JobsOfInterest)))
	return nil
}

func (g *GithubBridge) validateJob(ctx context.Context, job string) error {
	hooks, err := g.getWebHooks(ctx, job)
	if err != nil {
		return err
	}

	g.Log(fmt.Sprintf("%v -> %v hooks", job, len(hooks)))
	time.Sleep(time.Second * 5)

	if len(hooks) == 2 {
		for _, hook := range hooks {
			if strings.Contains(hook.Config.URL, "githubwebhook") {
				if len(hook.Events) != 7 {
					hook.Events = []string{"push", "issues", "create", "pull_request", "check_suite", "check_run", "status"}
					g.updateWebHook(ctx, job, hook)
				}
			}
		}
	}

	if len(hooks) == 1 {
		err := g.addWebHook(ctx, job, Webhook{
			Name:   "web",
			Active: true,
			Events: []string{"push", "issues", "create", "pull_request", "check_suite", "check_run", "status"},
			Config: Config{
				URL:         fmt.Sprintf("http://%v:50052/githubwebhook", g.config.ExternalIP),
				ContentType: "json",
				InsecureSSL: "1",
			}})
		if err != nil {
			return err
		}
	}

	return nil

}
