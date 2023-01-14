package github_client

import (
	"context"

	pb "github.com/brotherlogic/githubcard/proto"
	pbgs "github.com/brotherlogic/goserver"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GHClient struct {
	Gs        *pbgs.GoServer
	ErrorCode codes.Code
	Test      bool
	Issues    []*pb.Issue
}

func (c *GHClient) AddIssue(ctx context.Context, in *pb.Issue) (*pb.Issue, error) {
	if c.Test {
		c.Issues = append(c.Issues, in)
		if c.ErrorCode != codes.OK {
			return nil, status.Errorf(c.ErrorCode, "Forced Error")
		}
		return in, nil
	}
	conn, err := c.Gs.FDialServer(ctx, "githubcard")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := pb.NewGithubClient(conn)
	return client.AddIssue(ctx, in)
}

func (c *GHClient) GetIssues(ctx context.Context, req *pb.GetAllRequest) (*pb.GetAllResponse, error) {
	if c.Test {
		return &pb.GetAllResponse{Issues: c.Issues}, nil
	}

	conn, err := c.Gs.FDialServer(ctx, "githubcard")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := pb.NewGithubClient(conn)
	return client.GetAll(ctx, req)
}
