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
	if err != nil {
		log.Fatalf("Unable to dial: %v", err)
	}
	defer conn.Close()

	client := pb.NewGithubClient(conn)

	//	resp, err := client.Silence(ctx, &pb.SilenceRequest{Silence: "Crash for recordcollection", State: pb.SilenceRequest_UNSILENCE, Origin: "1569274842730506610"})
	//resp, err := client.Configure(ctx, &pb.ConfigureRequest{ExternalIp: os.Args[1]})
	resp, err := client.DeleteIssue(ctx, &pb.DeleteRequest{Issue: &pb.Issue{Number: 2593, Service: "home", Title: "Donkey", Body: "magic"}})
	fmt.Printf("%v -> %v\n", resp, err)

}
