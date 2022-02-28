package main

import (
	"testing"

	"golang.org/x/net/context"
)

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
