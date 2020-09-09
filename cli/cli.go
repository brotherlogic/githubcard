package main

import (
	"fmt"
	"log"

	"github.com/brotherlogic/goserver/utils"

	pb "github.com/brotherlogic/githubcard/proto"

	//Needed to pull in gzip encoding init
	_ "google.golang.org/grpc/encoding/gzip"
)

func main() {
	ctx, cancel := utils.BuildContext("githubcard-cli", "githubcard-cli")
	defer cancel()

	conn, err := utils.LFDialServer(ctx, "githubcard")
	defer conn.Close()

	if err != nil {
		log.Fatalf("Unable to dial: %v", err)
	}

	client := pb.NewGithubClient(conn)

	//	resp, err := client.Silence(ctx, &pb.SilenceRequest{Silence: "Crash for recordcollection", State: pb.SilenceRequest_UNSILENCE, Origin: "1569274842730506610"})
	resp, err := client.RegisterJob(ctx, &pb.RegisterRequest{Job: "recordscores"})
	fmt.Printf("%v and %v\n", resp, err)

}
