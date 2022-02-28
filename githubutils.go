package main

import (
	"fmt"
	"strings"

	"golang.org/x/net/context"
)

func (g *GithubBridge) validateJob(ctx context.Context, job string) error {
	hooks, err := g.getWebHooks(ctx, job)
	if err != nil {
		return err
	}

	if len(g.external) == 0 {
		return fmt.Errorf("No external IP registered")
	}

	if len(hooks) == 1 {
		for _, hook := range hooks {
			if strings.Contains(hook.Config.URL, "githubwebhook") {
				if len(hook.Events) != 7 {
					hook.Events = []string{"push", "issues", "create", "pull_request", "check_suite", "check_run", "status"}
					hook.Config.Secret = g.githubsecret
					g.updateWebHook(ctx, job, hook)
				}
				if len(hook.Config.Secret) == 0 {
					hook.Config.Secret = g.githubsecret
					g.Log(fmt.Sprintf("Setting secret for %v", job))
					g.updateWebHook(ctx, job, hook)
				}
				if !strings.Contains(hook.Config.URL, g.external) {
					hook.Config.URL = fmt.Sprintf("http://%v:50052/githubwebhook", g.external)
					hook.Config.Secret = g.githubsecret
					g.updateWebHook(ctx, job, hook)
				}

				if hook.Config.ContentType != "json" {
					hook.Config.ContentType = "json"
					hook.Config.Secret = g.githubsecret
					g.updateWebHook(ctx, job, hook)
				}
			}
		}
	}

	if len(hooks) == 0 {
		hook := Webhook{
			Name:   "web",
			Active: true,
			Events: []string{"push", "issues", "create", "pull_request", "check_suite", "check_run", "status"},
			Config: Config{
				URL:         fmt.Sprintf("http://%v:50052/githubwebhook", g.external),
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
