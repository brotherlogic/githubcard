package main

import (
	"log"
	"testing"

	pb "github.com/brotherlogic/githubcard/proto"
	"golang.org/x/net/context"
)

func TestProcSticky(t *testing.T) {
	log.Printf("TestProcSticky")
	g := InitTest()
	g.issues = append(g.issues, &pb.Issue{Service: "blah", Title: "blah", Body: "blah"})
	g.procSticky(context.Background())

	if len(g.issues) != 0 {
		t.Errorf("Issue was not added: %v", g.issues)
	}
}

func TestProcStickyfail(t *testing.T) {
	log.Printf("TestProcSticky")
	g := InitTest()
	g.procSticky(context.Background())

	if len(g.issues) != 0 {
		t.Errorf("Issue was not added: %v", g.issues)
	}
}

func TestValidateJobs(t *testing.T) {
	s := InitTest()
	s.config.JobsOfInterest = append(s.config.JobsOfInterest, "crasher")
	err := s.validateJobs(context.Background())

	if err != nil {
		t.Errorf("Validation failed")
	}
}

func TestValidateJobsFail(t *testing.T) {
	s := InitTest()
	s.config.JobsOfInterest = append(s.config.JobsOfInterest, "madeupjob")
	err := s.validateJobs(context.Background())

	if err == nil {
		t.Errorf("Validation did not fail")
	}
}
