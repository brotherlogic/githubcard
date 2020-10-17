package main

import (
	"fmt"
	"strings"
	"time"

	"golang.org/x/net/context"
)

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
	if job == "recordscores" {
		g.Log(fmt.Sprintf("FOUND %+v and %v -> %v", hooks, err, len(hooks)))
		time.Sleep(time.Second * 2)
	}
	if err != nil {
		return err
	}

	if len(hooks) == 2 {
		for _, hook := range hooks {
			if strings.Contains(hook.Config.URL, "githubwebhook") {
				if len(hook.Events) != 7 {
					hook.Events = []string{"push", "issues", "create", "pull_request", "check_suite", "check_run", "status"}
					g.updateWebHook(ctx, job, hook)
				}
				if len(hook.Config.Secret) == 0 {
					hook.Config.Secret = g.githubsecret
					g.Log(fmt.Sprintf("Setting secret for %v", job))
					g.updateWebHook(ctx, job, hook)
				}
				if !strings.Contains(hook.Config.URL, g.config.ExternalIP) {
					hook.Config.URL = fmt.Sprintf("http://%v:50052/githubwebhook", g.config.ExternalIP)
					g.updateWebHook(ctx, job, hook)
				}

				if hook.Config.ContentType != "json" {
					hook.Config.ContentType = "json"
					g.updateWebHook(ctx, job, hook)
				}
			}
		}
	}

	if len(hooks) == 1 {
		hook := Webhook{
			Name:   "web",
			Active: true,
			Events: []string{"push", "issues", "create", "pull_request", "check_suite", "check_run", "status"},
			Config: Config{
				URL:         fmt.Sprintf("http://%v:50052/githubwebhook", g.config.ExternalIP),
				ContentType: "json",
				InsecureSSL: "1",
			}}
		g.Log(fmt.Sprintf("Adding %v", hook))
		err := g.addWebHook(ctx, job, hook)
		if err != nil {
			return err
		}
	}

	return nil

}
