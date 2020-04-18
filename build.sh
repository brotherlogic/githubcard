protoc -I=./proto --go_out=plugins=grpc:./proto proto/githubcard.proto
mv proto/github.com/brotherlogic/githubcard/proto/* ./proto
