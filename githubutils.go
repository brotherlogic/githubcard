package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	pbgh "github.com/brotherlogic/githubcard/proto"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/context"
)

var (
	PORT_NUMBER = 50051
)

func (g *GithubBridge) metrics(config *pbgh.Config) {
	counts := make(map[string]int32)
	for _, issue := range config.GetIssues() {
		counts[issue.GetService()]++
	}

	for service, count := range counts {
		issues.With(prometheus.Labels{"service": service}).Set(float64(count))
	}
}
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
		g.CtxLog(ctx, fmt.Sprintf("Adding %v", hook))
		err := g.addWebHook(ctx, job, hook)
		if err != nil {
			return err
		}
	}

	return nil
}

type issueReturn struct {
	Url         string
	Title       string
	Body        string
	CreatedAt   string      `json:"created_at"`
	PullRequest PullRequest `json:"pull_request"`
}

type PullRequest struct {
	Url string
}

// GetIssues Gets github issues for a given project
func (b *GithubBridge) GetIssues(ctx context.Context) ([]*pbgh.Issue, error) {
	urlv := "https://api.github.com/issues?state=open&filter=all"
	body, _, err := b.visitURL(ctx, urlv)

	if err != nil {
		return nil, err
	}

	var issuesRet []*issueReturn
	err = json.Unmarshal([]byte(body), &issuesRet)

	var issues []*pbgh.Issue
	for _, ir := range issuesRet {
		splits := strings.Split(ir.Url, "/")
		service := splits[5]
		number, _ := strconv.ParseInt(splits[7], 10, 32)
		date, err := time.Parse(time.RFC3339, ir.CreatedAt)
		if err != nil {
			return nil, err
		}
		issues = append(issues, &pbgh.Issue{
			Service:   service,
			Number:    int32(number),
			Title:     ir.Title,
			Body:      ir.Body,
			DateAdded: date.Unix()})
	}

	return issues, err
}
