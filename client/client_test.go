package github_client

import (
	"context"
	"testing"

	pb "github.com/brotherlogic/githubcard/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestClientAddFail(t *testing.T) {
	s := &GHClient{Test: true}
	s.AddErrorCode = codes.DataLoss

	res, err := s.AddIssue(context.Background(), &pb.Issue{})
	if status.Code(err) != codes.DataLoss {
		t.Errorf("This should have failed: %v / %v", res, err)
	}
}
