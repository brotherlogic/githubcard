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
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/net/context"
)

var (
	PORT_NUMBER = 50051

	outstanding = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "githubcard_outstanding",
		Help: "The number of issues added per binary",
	}, []string{"job", "type"})
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

	prs, _, err := g.client.PullRequests.List(ctx, "brotherlogic", job, &github.PullRequestListOptions{})
	if err != nil {
		return err
	}
	outstanding.With(prometheus.Labels{"job": job, "type": "pull"}).Set(float64(len(prs)))

	brs, _, err := g.client.Repositories.ListBranches(ctx, "brotherlogic", job, &github.BranchListOptions{})
	if err != nil {
		return err
	}
	outstanding.With(prometheus.Labels{"job": job, "type": "branch"}).Set(float64(len(brs)))

	for _, hook := range hooks {
		if strings.Contains(hook.Config.URL, "travis") {
			g.deleteWebHook(ctx, job, hook)
		}
		if strings.Contains(hook.Config.URL, "githubwebhook") {
			if len(hook.Events) != 7 {
				hook.Events = []string{"push", "issues", "create", "pull_request", "check_suite", "check_run", "status"}
				hook.Config.Secret = g.githubsecret
				g.updateWebHook(ctx, job, hook)
			}
			if len(hook.Config.Secret) == 0 {
				hook.Config.Secret = g.githubsecret
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
		hook.Config.Secret = g.githubsecret
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
		if resp.StatusCode == 401 {
			g.CtxLog(ctx, fmt.Sprintf("401 error from github: %v", g.accessCode))
		}
		return err
	}

	// Add a main branch if there isn't one

	nrepo := &github.Repository{}
	updated := false
	if repo.GetDefaultBranch() != "main" {
		db := "main"
		nrepo.DefaultBranch = &db
		updated = true

		// Push a new commit
		_, _, err := g.client.Repositories.RenameBranch(ctx, "brotherlogic", job, "master", "main")
		if err != nil {
			g.RaiseIssue("Bad branch rename", fmt.Sprintf("Bad rename: %v", err))
		}
	}

	if !repo.GetDeleteBranchOnMerge() {
		deleteBranch := true
		nrepo.DeleteBranchOnMerge = &deleteBranch
		updated = true
	}

	if updated {
		g.client.Repositories.Edit(ctx, "brotherlogic", job, nrepo)
	}

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
	_, err = g.client.Actions.CreateOrUpdateRepoSecret(ctx, "brotherlogic", job, secret)
	if err != nil {
		return err
	}

	if !repo.GetPrivate() {
		bp, resp, err := g.client.Repositories.GetBranchProtection(ctx, "brotherlogic", job, "main")
		if err != nil {
			g.RaiseIssue("Bad branch pull", fmt.Sprintf("For %v Got %v and %+v", job, err, resp))
		}
		_, resp, err = g.client.Repositories.UpdateBranchProtection(ctx, "brotherlogic", job, "main",
			&github.ProtectionRequest{
				EnforceAdmins: true,
				RequiredStatusChecks: &github.RequiredStatusChecks{
					Checks: []*github.RequiredStatusCheck{
						{Context: "basic_assess"},
					},
				},
			})
		if err != nil {
			g.RaiseIssue("Bad Branch Update", fmt.Sprintf("For %v Got %v and %v", job, err, resp))
		}

		foundBasicAssess := false
		if bp != nil && bp.GetRequiredStatusChecks() != nil && bp.GetRequiredStatusChecks().Checks != nil {
			for _, check := range bp.GetRequiredStatusChecks().Checks {
				if check.Context == "basic_assess" {
					foundBasicAssess = true
				}
			}
		}

		if !foundBasicAssess {
			g.RaiseIssue("Missing basic assess", fmt.Sprintf("Did not find basic assess for %v: %+v", job, bp))
		}

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
