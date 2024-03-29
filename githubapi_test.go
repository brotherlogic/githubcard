package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"

	keystoreclient "github.com/brotherlogic/keystore/client"
	github "github.com/google/go-github/v50/github"
	mock "github.com/migueleliasweb/go-github-mock/src/mock"

	pb "github.com/brotherlogic/githubcard/proto"
)

func InitTest() *GithubBridge {
	s := Init()
	s.getter = testFileGetter{}
	s.accessCode = "token"
	s.SkipLog = true
	s.SkipIssue = true
	s.GoServer.KSclient = *keystoreclient.GetTestClient(".test")

	resp := "Le0rUuTnG6ACQS9dfUTzSfHzFL2b+lCdDCHyQRplMGE="
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposActionsSecretsByOwnerByRepo,
			github.Secrets{},
		),

		mock.WithRequestMatch(
			mock.GetReposActionsSecretsPublicKeyByOwnerByRepo,
			github.PublicKey{
				Key:   &resp,
				KeyID: &resp,
			},
		),

		mock.WithRequestMatch(
			mock.PutReposActionsSecretsByOwnerByRepoBySecretName,
			"",
		),

		mock.WithRequestMatch(
			mock.GetReposByOwnerByRepo,
			github.Repository{},
		),
	)
	s.client = github.NewClient(mockedHTTPClient)

	return s
}

type failGetter struct{}

func (httpGetter failGetter) Post(url string, data string) (*http.Response, error) {
	return nil, errors.New("Built to Fail")
}

func (httpGetter failGetter) Patch(url string, data string) (*http.Response, error) {
	return nil, errors.New("Built to Fail")
}

func (httpGetter failGetter) Put(url string, data string) (*http.Response, error) {
	return nil, errors.New("Built to Fail")
}

func (httpGetter failGetter) Get(ctx context.Context, url string) (*http.Response, error) {
	return nil, errors.New("Built to Fail")
}

func (httpGetter failGetter) Delete(url string) (*http.Response, error) {
	return nil, errors.New("Built to Fail")
}

type testFileGetter struct {
	jsonBreak bool
	failPost  bool
}

func (httpGetter testFileGetter) Post(url string, data string) (*http.Response, error) {
	log.Printf("url  %v", url)
	log.Printf("data %v", data)
	response := &http.Response{}
	if httpGetter.failPost {
		return response, fmt.Errorf("Built to fail")
	}
	strippedURL := strings.Replace(strings.Replace(url[22:], "?", "_", -1), "&", "_", -1)
	if httpGetter.jsonBreak {
		strippedURL = strings.Replace(strippedURL, "brotherlogic", "broke", -1)
	}
	blah, err := os.Open("testdata" + strippedURL + "_access_token=token")
	if err != nil {
		blah, err = os.Open("testdata" + strippedURL + "_")
		if err != nil {
			log.Printf("Error patching test file %v", err)
		}
	}
	response.Body = blah
	return response, nil
}

func (httpGetter testFileGetter) Patch(url string, data string) (*http.Response, error) {
	log.Printf("url  %v", url)
	log.Printf("data %v", data)
	response := &http.Response{}
	if httpGetter.failPost {
		return response, fmt.Errorf("Built to fail")
	}
	strippedURL := strings.Replace(strings.Replace(url[22:], "?", "_", -1), "&", "_", -1)
	if httpGetter.jsonBreak {
		strippedURL = strings.Replace(strippedURL, "token", "broke", -1)
	}
	blah, err := os.Open("testdata" + strippedURL + "_access_token=token")
	if err != nil {
		blah, err = os.Open("testdata" + strippedURL + "_")
		if err != nil {
			log.Printf("Error patching test file %v", err)
		}
	}
	response.Body = blah
	return response, nil
}

func (httpGetter testFileGetter) Put(url string, data string) (*http.Response, error) {
	log.Printf("url  %v", url)
	log.Printf("data %v", data)
	response := &http.Response{}
	if httpGetter.failPost {
		return response, fmt.Errorf("Built to fail")
	}
	strippedURL := strings.Replace(strings.Replace(url[22:], "?", "_", -1), "&", "_", -1)
	if httpGetter.jsonBreak {
		strippedURL = strings.Replace(strippedURL, "token", "broke", -1)
	}
	blah, err := os.Open("testdata" + strippedURL + "_access_token=token")
	if err != nil {
		blah, err = os.Open("testdata" + strippedURL + "_")
		if err != nil {
			log.Printf("Error patching test file %v", err)
		}
	}
	response.Body = blah
	response.StatusCode = 200
	return response, nil
}

func (httpGetter testFileGetter) Get(ctx context.Context, url string) (*http.Response, error) {
	response := &http.Response{}
	strippedURL := strings.Replace(strings.Replace(url[22:], "?", "_", -1), "&", "_", -1)
	blah, err := os.Open("testdata" + strippedURL + "_access_token=token")
	if err != nil {
		blah, err = os.Open("testdata" + strippedURL + "_")
		if err != nil {
			log.Printf("Error opening test file %v", err)
		}
	}
	response.Body = blah
	return response, nil
}

func (httpGetter testFileGetter) Delete(url string) (*http.Response, error) {
	response := &http.Response{}
	strippedURL := strings.Replace(strings.Replace(url[22:], "?", "_", -1), "&", "_", -1)
	blah, err := os.Open("testdata" + strippedURL + "_access_token=token")
	if err != nil {
		log.Printf("Error opening test file %v", err)
	}
	response.Body = blah
	return response, nil
}

func TestAddBadSilence(t *testing.T) {
	s := InitTest()

	_, err := s.Silence(context.Background(), &pb.SilenceRequest{State: pb.SilenceRequest_SILENCE, Silence: "Unfinished call"})

	if err == nil {
		t.Errorf("Bad silence was not silenced")
	}
}

func TestGetIssue(t *testing.T) {
	log.Printf("TestGetIssue")
	s := InitTest()
	ib, err := s.Get(context.Background(), &pb.Issue{Service: "Home", Number: 12})

	if err != nil {
		t.Fatalf("Error in getting issue: %v", err)
	}

	if ib.Number != 12 {
		t.Errorf("Issue has not been returned correctly: %v", ib)
	}
}

func TestCreatePullRequesrt(t *testing.T) {
	s := InitTest()
	s.CreatePullRequest(context.Background(), &pb.PullRequest{Job: "blah", Branch: "blah"})
}

func TestAddMilestone(t *testing.T) {
	s := InitTest()
	num, err := s.AddMilestone(context.Background(), &pb.AddMilestoneRequest{Title: "test", Description: "Testing", Repo: "frametracker"})
	if err != nil {
		t.Errorf("Bad add milestone: %v", err)
	}

	if num.GetNumber() != int32(1) {
		t.Errorf("Bad number: %v", num.GetNumber())
	}
}

func TestUpdateMilestone(t *testing.T) {
	s := InitTest()
	_, err := s.UpdateMilestone(context.Background(), &pb.UpdateMilestoneRequest{Number: 1, Repo: "frametracker", State: "closed"})
	if err != nil {
		t.Errorf("Bad update milesonte: %v", err)
	}
}

func TestClosePullRequesrt(t *testing.T) {
	s := InitTest()
	resp, err := s.ClosePullRequest(context.Background(), &pb.CloseRequest{Job: "frametracker", PullNumber: 16, Sha: "f4256902623ce71c7dbcd02f5c3a959afbd7e395", BranchName: "testbranch"})
	if err != nil {
		t.Errorf("Bad pr %v and %v", resp, err)
	}
}

func TestCloseMadeUpPullRequest(t *testing.T) {
	s := InitTest()
	s.getter = &testFileGetter{failPost: true}
	resp, err := s.ClosePullRequest(context.Background(), &pb.CloseRequest{Job: "madeup", PullNumber: 16, Sha: "f4256902623ce71c7dbcd02f5c3a959afbd7e395", BranchName: "testbranch"})
	if err == nil {
		t.Errorf("PR was fine: %v", resp)
	}
}

func TestGetPullRequest(t *testing.T) {
	s := InitTest()
	pull, err := s.GetPullRequest(context.Background(), &pb.PullRequest{Job: "githubreceiver", PullNumber: 24})
	if err != nil {
		t.Fatalf("Error getting pull request: %v", err)
	}

	if pull.NumberOfCommits != 7 {
		t.Errorf("Wrong number of commits returend: %v", pull)
	}

	if !pull.IsOpen {
		t.Errorf("Pull request should be open %v", pull)
	}
}
