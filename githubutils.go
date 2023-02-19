package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"crypto/rand"
    "encoding/base64"
    "golang.org/x/crypto/nacl/box"

	github "github.com/google/go-github/v50/github"

	pbgh "github.com/brotherlogic/githubcard/proto"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

	g.CtxLog(ctx, fmt.Sprintf("Checking for head branches"))
	// Ensure that we delete head branches
	repo, err := g.getRepo(ctx, job)
	if err != nil {
		return err
	}
	if !repo.DeleteBranchOnMerge {
		err := g.updateRepo(ctx, job, true)
		if err != nil {
			return err
		}
	}

	// Enable branch protection
	prot, err := g.getBranchProtection(ctx, job, "main")
	if err != nil && status.Code(err) != codes.NotFound {
		return err
	}
	if status.Code(err) == codes.NotFound {
		err := g.updateBranchProtection(ctx, job, &BranchProtection{Url: fmt.Sprintf("https://api.github.com/repos/brotherlogic/%v/branches/main/protection", job)})
		if err != nil {
			g.RaiseIssue("Bad branch protection update", fmt.Sprintf("%v is the error", err))
			return err
		}
	} else {
		foundChecks := 0
		for _, check := range prot.RequiredStatusChecks.Checks {
			if check.Context == "basic_assess" {
				foundChecks++
			}
		}
		if prot.RequiredPullRequestReviews.RequiredApprovingReviewCount != 0 || !prot.RequiredStatusChecks.Strict || foundChecks == 0 || !prot.EnforceAdmins.Enabled {
			g.RaiseIssue("Needed Pull Request", fmt.Sprintf("%v needed a pull request update for pull request required, %+v", job, prot))
			err := g.updateBranchProtection(ctx, job, &BranchProtection{
				Url:                        fmt.Sprintf("https://api.github.com/repos/brotherlogic/%v/branches/main/protection", job),
				RequiredPullRequestReviews: RequiredPullRequestReviews{RequiredApprovingReviewCount: 1},
				EnforceAdmins:              EnforceAdmins{Enabled: true},
				RequiredStatusChecks: RequiredStatusChecks{
					Strict: true,
					Checks: []Check{{Context: "basic_assess", AppId: -1}},
				},
			})
			if err != nil {
				return err
			}
		}
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
			KeyID: "PERSONAL_TOKEN",
			EncryptedValue: eval,
		}
		g.client.Actions.CreateOrUpdateRepoSecret(ctx, "brotherlogic", job,secret )
	}

	return nil
}


func encryptSecret(pk, secret string) (string, error) {
    var pkBytes [32]byte
    copy(pkBytes[:], pk)
    secretBytes := []byte(secret)

    out := make([]byte, 0,
        len(secretBytes)+
        box.Overhead+
        len(pkBytes))

    enc, err := box.SealAnonymous(
        out, secretBytes, &pkBytes, rand.Reader,
    )
    if err != nil {
        return "", err
    }

    encEnc := base64.StdEncoding.EncodeToString(enc)

    return encEnc, nil
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
