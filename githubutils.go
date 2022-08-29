package main

import (
	"fmt"
	"strings"

	"golang.org/x/net/context"
)

var (
	PORT_NUMBER = 50051
)

func (g *GithubBridge) validateJob(ctx context.Context, job string) error {
	hooks, err := g.getWebHooks(ctx, job)
	if err != nil {
		return err
	}

	if len(g.external) == 0 {
		return fmt.Errorf("No external IP registered")
	}

	for _, hook := range hooks {
		if strings.Contains(hook.Config.URL, "travis") {
			g.deleteWebHook(ctx, job, hook)
		}
		if strings.Contains(hook.Config.URL, "githubwebhook") {
			g.CtxLog(ctx, fmt.Sprintf("EV %v Sec %v Ext %v CT %v", len(hook.Events), len(hook.Config.Secret), hook.Config.URL, hook.Config.ContentType))
			if len(hook.Events) != 7 {
				hook.Events = []string{"push", "issues", "create", "pull_request", "check_suite", "check_run", "status"}
				hook.Config.Secret = g.githubsecret
				g.updateWebHook(ctx, job, hook)
			}
			if len(hook.Config.Secret) == 0 {
				hook.Config.Secret = g.githubsecret
				g.CtxLog(ctx, fmt.Sprintf("Setting secret for %v", job))
				g.updateWebHook(ctx, job, hook)
			}
			if !strings.Contains(hook.Config.URL, g.external) || !strings.Contains(hook.Config.URL, fmt.Sprintf("%v", PORT_NUMBER)) {
				hook.Config.URL = fmt.Sprintf("http://%v:%v/githubwebhook", g.external, PORT_NUMBER)
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

	if len(hooks) == 0 {
		hook := Webhook{
			Name:   "web",
			Active: true,
			Events: []string{"push", "issues", "create", "pull_request", "check_suite", "check_run", "status"},
			Config: Config{
				URL:         fmt.Sprintf("http://%v:%v/githubwebhook", g.external, PORT_NUMBER),
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
