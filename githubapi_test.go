package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/brotherlogic/keystore/client"

	pb "github.com/brotherlogic/githubcard/proto"
)

func InitTest() *GithubBridge {
	s := Init()
	s.getter = testFileGetter{}
	s.accessCode = "token"
	s.SkipLog = true
	s.GoServer.KSclient = *keystoreclient.GetTestClient(".test")
	return s
}

type failGetter struct{}

func (httpGetter failGetter) Post(url string, data string) (*http.Response, error) {
	return nil, errors.New("Built to Fail")
}

func (httpGetter failGetter) Get(url string) (*http.Response, error) {
	return nil, errors.New("Built to Fail")
}

type testFileGetter struct{ jsonBreak bool }

func (httpGetter testFileGetter) Post(url string, data string) (*http.Response, error) {
	log.Printf("url  %v", url)
	log.Printf("data %v", data)
	response := &http.Response{}
	strippedURL := strings.Replace(strings.Replace(url[22:], "?", "_", -1), "&", "_", -1)
	if httpGetter.jsonBreak {
		strippedURL = strings.Replace(strippedURL, "token", "broke", -1)
	}
	blah, err := os.Open("testdata" + strippedURL)
	if err != nil {
		log.Printf("Error opening test file %v", err)
	}
	response.Body = blah
	return response, nil
}

func (httpGetter testFileGetter) Get(url string) (*http.Response, error) {
	response := &http.Response{}
	strippedURL := strings.Replace(strings.Replace(url[22:], "?", "_", -1), "&", "_", -1)
	blah, err := os.Open("testdata" + strippedURL)
	if err != nil {
		log.Printf("Error opening test file %v", err)
	}
	response.Body = blah
	return response, nil
}

func TestAddIssue(t *testing.T) {
	log.Printf("TestAddIssue")
	issue := &pb.Issue{Title: "Testing", Body: "This is a test issue", Service: "Home"}

	s := InitTest()
	ib, err := s.AddIssue(context.Background(), issue)

	if err != nil {
		t.Fatalf("Error in adding issue: %v", err)
	}

	if ib.Number != 494 {
		t.Errorf("Issue has not been added: %v", ib.Number)
	}
}

func TestAddBlankIssue(t *testing.T) {
	s := InitTest()
	_, err := s.AddIssue(context.Background(), &pb.Issue{Title: "Long"})
	if err == nil {
		t.Errorf("Adding silenced issue did not fail")
	}

	if s.blankAlerts != 1 {
		t.Errorf("Number of blanks has not increased")
	}
}

func TestAddBadSilence(t *testing.T) {
	s := InitTest()

	_, err := s.Silence(context.Background(), &pb.SilenceRequest{State: pb.SilenceRequest_SILENCE, Silence: "Unfinished call"})

	if err == nil {
		t.Errorf("Bad silence was not silenced")
	}
}

func TestAddSilencedIssue(t *testing.T) {
	s := InitTest()

	_, err := s.AddIssue(context.Background(), &pb.Issue{Title: "Unfinished call", Body: "Blah", Service: "Home"})
	if err != nil {
		t.Errorf("We should be able to add this: %v", err)
	}

	_, err = s.Silence(context.Background(), &pb.SilenceRequest{State: pb.SilenceRequest_SILENCE, Silence: "Unfinished call", Origin: "blah"})
	if err != nil {
		t.Errorf("Unable to silence: %v", err)
	}

	_, err = s.AddIssue(context.Background(), &pb.Issue{Title: "Unfinished call", Body: "Blah", Service: "Home"})
	if err == nil {
		t.Errorf("We shouldn't have been able to add a silence issue")
	}

	if s.silencedAlerts != 1 {
		t.Errorf("Number of silences has not increased")
	}

	s.Silence(context.Background(), &pb.SilenceRequest{State: pb.SilenceRequest_UNSILENCE, Silence: "Unfinished call", Origin: "blah"})

	_, err = s.AddIssue(context.Background(), &pb.Issue{Title: "Unfinished call", Body: "Blah", Service: "Home", Sticky: true})
	if err != nil {
		t.Errorf("Unable to add issue once it's unsilenced: %v", err)
	}

	if s.silencedAlerts != 1 {
		t.Errorf("Number of silences has changed")
	}

	_, err = s.Silence(context.Background(), &pb.SilenceRequest{Silence: "Unfinished call", Origin: "blah"})
	if err == nil {
		t.Errorf("Malformed silence did not fail")
	}
}

func TestAddIssueToFakeService(t *testing.T) {
	log.Printf("TestAddIssueToFakeService")
	issue := &pb.Issue{Title: "Testing", Body: "This is a test issue", Service: "MadeUpService"}

	s := InitTest()
	_, err := s.AddIssue(context.Background(), issue)

	if err == nil {
		t.Fatalf("Error not added")
	}

	log.Printf("Error is %v", err)
}

func TestAddDoubleIssue(t *testing.T) {
	log.Printf("TestAddDoubleIssue")
	issue := &pb.Issue{Title: "Testing", Body: "This is a test issue", Service: "Home"}

	s := InitTest()
	ib, err := s.AddIssue(context.Background(), issue)

	if err != nil {
		t.Fatalf("Error in adding issue: %v", err)
	}

	if ib.Number != 494 {
		t.Errorf("Issue has not been added: %v", ib.Number)
	}

	_, err = s.AddIssue(context.Background(), issue)
	if err == nil {
		t.Errorf("Double add has not failed")
	}
}

func TestAddDoubleIssueWithSticky(t *testing.T) {
	log.Printf("TestAddDoubleIssueWithSticky")
	issue := &pb.Issue{Title: "Testing", Body: "This is a test issue", Service: "Home", Sticky: true}

	s := InitTest()
	ib, err := s.AddIssue(context.Background(), issue)

	if err != nil {
		t.Fatalf("Error in adding issue: %v", err)
	}

	if ib.Number != 494 {
		t.Errorf("Issue has not been added: %v", ib.Number)
	}

	_, err = s.AddIssue(context.Background(), issue)
	if err != nil {
		t.Errorf("Double add of sticky has failed: %v", err)
	}
}

func TestAddIssueFail(t *testing.T) {
	log.Printf("TestAddIssueFail")
	issue := &pb.Issue{Title: "Testing", Body: "This is a test issue", Service: "Home"}

	s := InitTest()
	s.getter = failGetter{}
	_, err := s.AddIssue(context.Background(), issue)

	if err == nil {
		t.Fatalf("No Error returned")
	}
}

func TestAddIssueFailWithSticky(t *testing.T) {
	log.Printf("TestAddIssueFailWithSticky")
	issue := &pb.Issue{Title: "Testing", Body: "This is a test issue", Service: "Home", Sticky: true}

	s := InitTest()
	s.getter = failGetter{}
	_, err := s.AddIssue(context.Background(), issue)

	if err != nil {
		t.Fatalf("Adding sticky issue has failed: %v", err)
	}
}

func TestAddIssueFJSONail(t *testing.T) {
	log.Printf("TestAddIssueFJSONail")
	issue := &pb.Issue{Title: "Testing", Body: "This is a test issue", Service: "Home"}

	s := InitTest()
	s.getter = testFileGetter{jsonBreak: true}
	_, err := s.AddIssue(context.Background(), issue)

	if err == nil {
		t.Fatalf("No Error returned")
	}
}

func TestSubmitComplexIssue(t *testing.T) {
	log.Printf("TestSubmitComplet")
	issue := &pb.Issue{Title: "CRASHER REPORT", Service: "crasher", Body: "2017/09/26 17:48:18 ip:\"192.168.86.28\" port:50057 name:\"crasher\" identifier:\"framethree\"  is Servingpanic: Whoopsiegoroutine 41 [running]:panic(0x3b13f8, 0x109643f8)\t/usr/lib/go-1.7/src/runtime/panic.go:500 +0x33cmain.crash()\t/home/simon/gobuild/src/github.com/brotherlogic/crasher/Crasher.go:36 +0x6ccreated by github.com/brotherlogic/goserver.(*GoServer).Serve\t/home/simon/gobuild/src/github.com/brotherlogic/goserver/goserverapi.go:126+0x254"}
	s := InitTest()
	ib, err := s.AddIssue(context.Background(), issue)

	if err != nil {
		t.Fatalf("Error in adding issue: %v", err)
	}

	if ib.Number != 15 {
		t.Errorf("Issue has not been added: %v", ib.Number)
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

func TestGetAllIssues(t *testing.T) {
	s := InitTest()
	_, err := s.GetAll(context.Background(), &pb.GetAllRequest{})
	if err == nil {
		t.Errorf("Get all did not fail")
	}
}

func TestValidateJobs(t *testing.T) {
	s := InitTest()
	err := s.validateJobs(context.Background())

	if err != nil {
		t.Errorf("Validation failed")
	}
}
