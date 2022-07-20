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

	pb "github.com/brotherlogic/githubcard/proto"
)

func InitTest() *GithubBridge {
	s := Init()
	s.getter = testFileGetter{}
	s.accessCode = "token"
	s.SkipLog = true
	s.GoServer.KSclient = *keystoreclient.GetTestClient(".test")
	s.config.ExternalIP = "Test"
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

func (httpGetter failGetter) Get(url string) (*http.Response, error) {
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
		log.Printf("Error opening test file %v", err)
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
		log.Printf("Error opening test file %v", err)
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
		log.Printf("Error opening test file %v", err)
	}
	response.Body = blah
	response.StatusCode = 200
	return response, nil
}

func (httpGetter testFileGetter) Get(url string) (*http.Response, error) {
	response := &http.Response{}
	strippedURL := strings.Replace(strings.Replace(url[22:], "?", "_", -1), "&", "_", -1)
	blah, err := os.Open("testdata" + strippedURL + "_access_token=token")
	if err != nil {
		log.Printf("Error opening test file %v", err)
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

func TestGetAllIssuesLatestWithNoEntries(t *testing.T) {
	s := InitTest()

	_, err := s.GetAll(context.Background(), &pb.GetAllRequest{LatestOnly: true})
	if err != nil {
		t.Errorf("Get all did fail: %v", err)
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

func TestConfigure(t *testing.T) {
	s := InitTest()
	s.Configure(context.Background(), &pb.ConfigureRequest{ExternalIp: "maic"})

	if s.config.ExternalIP != "maic" {
		t.Errorf("Wrong: %v", s.config)
	}
}
