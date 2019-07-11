package main

import (
	"fmt"
	"strings"

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
		err := g.validateJob(ctx, j)

		if err != nil {
			return err
		}
	}

	return nil
}

func (g *GithubBridge) validateJob(ctx context.Context, job string) error {
	hooks, err := g.getWebHooks(ctx, job)
	if err != nil {
		return err
	}

	g.Log(fmt.Sprintf("Found %+v webhooks", len(hooks)))

	if len(hooks) == 2 {
		for _, hook := range hooks {
			if strings.Contains(hook.Config.URL, "githubwebhook") {
				g.Log(fmt.Sprintf("Found %v events", len(hook.Events)))
				if len(hook.Events) != 6 {
					g.Log(fmt.Sprintf("Updating webhooks"))
					hook.Events = []string{"push", "issues", "create", "pull_request", "check_suite", "check_run"}
					g.updateWebHook(ctx, job, hook)
				}
			}
		}
	}

	if len(hooks) == 1 {
		err := g.addWebHook(ctx, job, Webhook{
			Name:   "web",
			Active: true,
			Events: []string{"push", "issues", "create", "pull_request"},
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
