package github_client

import (
	"context"

	pb "github.com/brotherlogic/githubcard/proto"
	pbgs "github.com/brotherlogic/goserver"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GHClient struct {
	Gs           *pbgs.GoServer
	ErrorCode    codes.Code
	Test         bool
	Issues       []*pb.Issue
	lastNumber   int32
	AddErrorCode codes.Code
}

func (c *GHClient) DeleteIssue(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	if c.Test {
		if c.ErrorCode != codes.OK {
			return nil, status.Errorf(c.ErrorCode, "Forced Error")
		}
		var nissues []*pb.Issue
		for _, issue := range c.Issues {
			if issue.GetService() != req.GetIssue().GetService() || issue.GetNumber() != req.GetIssue().GetNumber() {
				nissues = append(nissues, issue)
			}
		}

		c.Issues = nissues
		return &pb.DeleteResponse{}, nil
	}
	conn, err := c.Gs.FDialServer(ctx, "githubcard")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := pb.NewGithubClient(conn)
	return client.DeleteIssue(ctx, req)
}

func (c *GHClient) AddIssue(ctx context.Context, in *pb.Issue) (*pb.Issue, error) {
	if c.Test {
		if c.ErrorCode != codes.OK {
			return nil, status.Errorf(c.ErrorCode, "Forced Error")
		}
		if c.AddErrorCode != codes.OK {
			return nil, status.Errorf(c.AddErrorCode, "Forced Error")
		}
		c.Issues = append(c.Issues, in)
		c.lastNumber++
		in.Number = c.lastNumber
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
		if c.ErrorCode != codes.OK {
			return nil, status.Errorf(c.ErrorCode, "Built to fail")
		}
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
