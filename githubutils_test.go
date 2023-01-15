package main

import (
	"context"
	"testing"
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
}
