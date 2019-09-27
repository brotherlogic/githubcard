package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/brotherlogic/goserver/utils"
	"google.golang.org/grpc"

	pb "github.com/brotherlogic/githubcard/proto"

	//Needed to pull in gzip encoding init
	_ "google.golang.org/grpc/encoding/gzip"
)

func main() {
	host, port, err := utils.Resolve("githubcard", "githubcard-cli")
	if err != nil {
		log.Fatalf("Unable to reach organiser: %v", err)
	}
	conn, err := grpc.Dial(host+":"+strconv.Itoa(int(port)), grpc.WithInsecure())
	defer conn.Close()

	if err != nil {
		log.Fatalf("Unable to dial: %v", err)
	}

	client := pb.NewGithubClient(conn)
	ctx, cancel := utils.BuildContext("githubcard-cli", "githubcard-cli")
	defer cancel()

	resp, err := client.Silence(ctx, &pb.SilenceRequest{Silence: "Crash for recordcollection", State: pb.SilenceRequest_UNSILENCE, Origin: "1569274842730506610"})
	fmt.Printf("%v and %v\n", resp, err)

}
