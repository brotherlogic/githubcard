package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jefflinse/githubsecret"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	github "github.com/google/go-github/v50/github"

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
		if issue.State != pbgh.Issue_CLOSED {
			counts[issue.GetService()]++
		}
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
	config, err := g.readIssues(ctx)

	if len(config.GetExternalIP()) == 0 {
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

	// Enable branch protection
	repo, resp, err := g.client.Repositories.Get(ctx, "brotherlogic", job)
	if err != nil {
		if resp.StatusCode == 403 {
			return status.Errorf(codes.ResourceExhausted, "Bad pull: %v", err)
		}
		return err
	}

	if repo.GetDefaultBranch() != "main" {
		g.RaiseIssue("Default Branch Change Needed", fmt.Sprintf("%v needs to change the default branch", job))
	}

	// Handle secrets
	secrets, resp, err := g.client.Actions.ListRepoSecrets(ctx, "brotherlogic", job, nil)
	if resp != nil {
		clientReads.Set(float64(resp.Rate.Remaining))
	}
	if err != nil {
		return err
	}
	found := false
	for _, secret := range secrets.Secrets {
		if secret.Name == "PERSONAL_TOKEN" {
			found = true
		}
	}

	g.CtxLog(ctx, fmt.Sprintf("SECRET: Found personal token (%v)", found))
	if !found {
		key, _, err := g.client.Actions.GetRepoPublicKey(ctx, "brotherlogic", job)
		if err != nil {
			return err
		}

		eval, err := encryptSecret(*key.Key, g.accessCode)
		if err != nil {
			return err
		}
		secret := &github.EncryptedSecret{
			Name:           "PERSONAL_TOKEN",
			EncryptedValue: eval,
			KeyID:          *key.KeyID,
		}
		bal, err := g.client.Actions.CreateOrUpdateRepoSecret(ctx, "brotherlogic", job, secret)
		if err != nil {
			return err
		}
		g.CtxLog(ctx, fmt.Sprintf("Added secret %+v -> %v,%v (%v)", secret, bal, err, g.accessCode))
	}

	return nil
}

func encryptSecret(pk, secret string) (string, error) {
	return githubsecret.Encrypt(pk, secret)
}

type RepoReturn struct {
	DeleteBranchOnMerge bool `json:"delete_branch_on_merge"`
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
		if ir.PullRequest.Url != "" {
			continue
		}
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
