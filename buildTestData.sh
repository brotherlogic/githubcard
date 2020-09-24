mkdir -p testdata/repos/brotherlogic/Home/issues/
mkdir -p testdata/repos/brotherlogic/MadeUpService/issues/
mkdir -p testdata/repos/brotherlogic/crasher
mkdir -p testdata/repos/brotherlogic/goserver
mkdir -p testdata/repos/brotherlogic/githubreceiver/pulls/24/
mkdir -p testdata/repos/brotherlogic/frametracker/pulls/16/
mkdir -p testdata/repos/brotherlogic/frametracker/milestones/1/
mkdir -p testdata/repos/brotherlogic/pullrequester/git/refs/heads/
mkdir -p testdata/user/

sleep 1
curl -H "Content-Type: application/json" --user-agent "GithubAgent" "https://api.github.com/repos/brotherlogic/goserver/hooks?access_token=$1"  > testdata/repos/brotherlogic/goserver/hooks_access_token=token
exit


sleep 1
curl -u "brotherlogic:$1" -H "Content-Type: application/json" --user-agent "GithubAgent" "https://api.github.com/user/issues"  > testdata/user/issues
exit

sleep 1
curl -u "brotherlogic:$1"-X POST -H "Content-Type: application/json" --user-agent "GithubAgent" "https://api.github.com/repos/brotherlogic/crasher/issues" -d '{"title": "Crash Report", "body": "2017/09/26 17:48:18 ip:\"192.168.86.28\" port:50057 name:\"crasher\" identifier:\"framethree\"  is Servingpanic: Whoopsiegoroutine 41 [running]:panic(0x3b13f8, 0x109643f8)\t/usr/lib/go-1.7/src/runtime/panic.go:500 +0x33cmain.crash()\t/home/simon/gobuild/src/github.com/brotherlogic/crasher/Crasher.go:36 +0x6ccreated by github.com/brotherlogic/goserver.(*GoServer).Serve\t/home/simon/gobuild/src/github.com/brotherlogic/goserver/goserverapi.go:126+0x254", "assignee": "brotherlogic"}' > testdata/repos/brotherlogic/crasher/issues
exit

sleep 1
curl -X PATCH -H "Content-Type: application/json" --user-agent "GithubAgent" "https://api.github.com/repos/brotherlogic/frametracker/milestones/1?access_token=$1" -d '{"state": "closed"}' > testdata/repos/brotherlogic/frametracker/milestones/1_access_token=token
exit


sleep 1
curl -X POST -H "Content-Type: application/json" --user-agent "GithubAgent" "https://api.github.com/repos/brotherlogic/frametracker/milestones?access_token=$1" -d '{"title": "Testing", "description": "This is a test issue"}' > testdata/repos/brotherlogic/frametracker/milestones_access_token=token
exit


sleep 1
curl -X DELETE -H "Content-Type: application/json" --user-agent "GithubAgent" "https://api.github.com/repos/brotherlogic/pullrequester/git/refs/heads/update_on_update?access_token=$1"  > testdata/repos/brotherlogic/pullrequester/git/refs/heads/update_on_update_access_token=token
exit


sleep 1
curl -X PUT -H "Content-Type: application/json" --user-agent "GithubAgent" "https://api.github.com/repos/brotherlogic/frametracker/pulls/16/merge?access_token=$1" -d '{"sha": "f4256902623ce71c7dbcd02f5c3a959afbd7e395"'} > testdata/repos/brotherlogic/frametracker/pulls/16/merge_access_token=token
exit


sleep 1
curl -H "Content-Type: application/json" --user-agent "GithubAgent" "https://api.github.com/repos/brotherlogic/githubreceiver/pulls/24?access_token=$1"  > testdata/repos/brotherlogic/githubreceiver/pulls/24_access_token=token
exit


sleep 1
curl -H "Content-Type: application/json" --user-agent "GithubAgent" "https://api.github.com/repos/brotherlogic/githubreceiver/pulls/24/commits?access_token=$1"  > testdata/repos/brotherlogic/githubreceiver/pulls/24/commits_access_token=token
exit


sleep 1
curl -H "Content-Type: application/json" --user-agent "GithubAgent" "https://api.github.com/repos/brotherlogic/crasher/hooks?access_token=$1"  > testdata/repos/brotherlogic/crasher/hooks_access_token=token
exit

sleep 1
curl -X POST -H "Content-Type: application/json" --user-agent "GithubAgent" "https://api.github.com/repos/brotherlogic/MadeUpService/issues?access_token=$1" -d '{"title": "Testing", "body": "This is a test issue", "assignee": "brotherlogic"}' > testdata/repos/brotherlogic/MadeUpService/issues_access_token=token
exit

sleep 1
curl -H "Content-Type: application/json" --user-agent "GithubAgent" "https://api.github.com/user/issues?access_token=$1"  > testdata/user/issues_access_token=token
exit

sleep 1
curl -X POST -H "Content-Type: application/json" --user-agent "GithubAgent" "https://api.github.com/repos/brotherlogic/crasher/issues?access_token=$1" -d '{"title": "CRASH REPORT", "body": "Starting Scan\n\n\n\npanic: rpc error: code = Unavailable desc = grpc: the connection is unavailablegoroutine 9 [running]:panic(0x3ddea0, 0x10bd8d20)\t/usr/lib/go-1.7/src/runtime/panic.go:500 +0x33cmain.Server.processCard(0x1090a680, 0x101, 0x0)\t/home/simon/gobuild/src/github.com/brotherlogic/recordgetter/recordget.go:242 +0x2acmain.Server.runSingle(0x1090a680, 0x101)\t/home/simon/gobuild/src/github.com/brotherlogic/recordgetter/recordget.go:300 +0x6cmain.Server.GetRecords(0x1090a680, 0x101)\t/home/simon/gobuild/src/github.com/brotherlogic/recordgetter/recordget.go:292 +0x100main.(Server).GetRecords-fm()\t/home/simon/gobuild/src/github.com/brotherlogic/recordgetter/recordget.go:373 +0x38created by github.com/brotherlogic/goserver.(*GoServer).Serve\t/home/simon/gobuild/src/github.com/brotherlogic/goserver/goserverapi.go:126 +0x254Finishing Scan\nFound Dead When Running", "assignee": "brotherlogic"}' > testdata/repos/brotherlogic/crasher/issues_access_token=token
exit

sleep 1
curl -X POST -H "Content-Type: application/json" --user-agent "GithubAgent" "https://api.github.com/repos/brotherlogic/Home/issues?access_token=$1" -d '{"title": "Testing", "body": "This is a test issue", "assignee": "brotherlogic"}' > testdata/repos/brotherlogic/Home/issues_access_token=token

sleep 1
curl -H "Content-Type: application/json" --user-agent "GithubAgent" "https://api.github.com/repos/brotherlogic/Home/issues/12?access_token=$1"  > testdata/repos/brotherlogic/Home/issues/12_access_token=token

sleep 1
curl -X POST -H "Content-Type: application/json" --user-agent "GithubAgent" "https://api.github.com/repos/brotherlogic/crasher/issues?access_token=$1" -d '{"title": "Crash Report", "body": "2017/09/26 17:48:18 ip:\"192.168.86.28\" port:50057 name:\"crasher\" identifier:\"framethree\"  is Servingpanic: Whoopsiegoroutine 41 [running]:panic(0x3b13f8, 0x109643f8)\t/usr/lib/go-1.7/src/runtime/panic.go:500 +0x33cmain.crash()\t/home/simon/gobuild/src/github.com/brotherlogic/crasher/Crasher.go:36 +0x6ccreated by github.com/brotherlogic/goserver.(*GoServer).Serve\t/home/simon/gobuild/src/github.com/brotherlogic/goserver/goserverapi.go:126+0x254", "assignee": "brotherlogic"}' > testdata/repos/brotherlogic/crasher/issues_access_token=token

sleep 1
curl -X POST -H "Content-Type: application/json" --user-agent "GithubAgent" "https://api.github.com/repos/brotherlogic/crasher/issues?access_token=$1" -d '{"title": "CRASH REPORT", "body": "Starting Scan\npanic: rpc error: code = Unavailable desc = grpc: the connection is unavailablegoroutine 9 [running]:panic(0x3ddea0, 0x10bd8d20)\t/usr/lib/go-1.7/src/runtime/panic.go:500 +0x33cmain.Server.processCard(0x1090a680, 0x101, 0x0)\t/home/simon/gobuild/src/github.com/brotherlogic/recordgetter/recordget.go:242 +0x2acmain.Server.runSingle(0x1090a680, 0x101)\t/home/simon/gobuild/src/github.com/brotherlogic/recordgetter/recordget.go:300 +0x6cmain.Server.GetRecords(0x1090a680, 0x101)\t/home/simon/gobuild/src/github.com/brotherlogic/recordgetter/recordget.go:292 +0x100main.(Server).GetRecords-fm()\t/home/simon/gobuild/src/github.com/brotherlogic/recordgetter/recordget.go:373 +0x38created by github.com/brotherlogic/goserver.(*GoServer).Serve\t/home/simon/gobuild/src/github.com/brotherlogic/goserver/goserverapi.go:126 +0x254Finishing Scan\nFound Dead When Running", "assignee": "brotherlogic"}' > testdata/repos/brotherlogic/crasher/issues_access_token=token
