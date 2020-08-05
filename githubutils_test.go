package main

import (
	"testing"

	"golang.org/x/net/context"
)

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

func TestValidateSingleJob(t *testing.T) {
	s := InitTest()
	err := s.validateJob(context.Background(), "crasher2")

	if err != nil {
		t.Errorf("Validation failed: %v", err)
	}
}

func TestValidatePostFailSingleJob(t *testing.T) {
	s := InitTest()
	s.getter = testFileGetter{failPost: true}
	err := s.validateJob(context.Background(), "crasher2")

	if err == nil {
		t.Errorf("Validation failed: %v", err)
	}
}

func TestValidatePostUpdate(t *testing.T) {
	s := InitTest()
	err := s.validateJob(context.Background(), "crasher3")

	if err != nil {
		t.Errorf("Validation failed: %v", err)
	}
}
