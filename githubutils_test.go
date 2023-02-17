package main

import (
	"context"
	"testing"

	pb "github.com/brotherlogic/githubcard/proto"
)

func TestReadIssues(t *testing.T) {
	s := InitTest()

	issues, err := s.GetIssues(context.Background())
	if err != nil {
		t.Errorf("What: %v", err)
	}

	if len(issues) == 0 {
		t.Fatalf("No issues returned")
	}

	if issues[0].Title == "" || issues[0].Number == 0 {
		t.Errorf("Bad read on issue: %v", issues[0])
	}

	if issues[0].DateAdded == 0 {
		t.Errorf("We haven't read the date: %v", issues[0])
	}

	if len(issues) != 1 {
		t.Errorf("Pull request counted as issue: %v", len(issues))
	}
}

func TestValidate(t *testing.T) {
	s := InitTest()
	_, err := s.Configure(context.Background(), &pb.ConfigureRequest{ExternalIp: "brotherlogic-backend.com"})
	if err != nil {
		t.Fatalf("Bad Configure: %v", err)
	}

	err = s.validateJob(context.Background(), "crasher")
	if err != nil {
		t.Errorf("Bad validate: %v", err)
	}
}
