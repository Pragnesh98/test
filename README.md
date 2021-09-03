# asterisk-ari

Asterisk ARI is service/wrapper around [Asterisk](https://github.com/asterisk/asterisk) which helps to submit the operations like Play, Record, Create Channel, Answer etc to Asterisk. This actually is a bot feature implementation where we listen to the user's response, record it, convert this speech to text, and hit YM bot API to get the response.

#### Prerequisite:
- System should have Asterisk and underlying modules installed
- If this is being built locally on the server, server should have golang installed and setup
- [protoc](http://google.github.io/proto-lens/installing-protoc.html) and [protoc-gen-go](https://pkg.go.dev/github.com/golang/protobuf/protoc-gen-go)

#### Build Steps:
```sh
$ cd asterisk-ari
$ cd utils/grpc/proto && protoc --go_out=plugins=grpc:. *.proto
```
Run the following in the root directory of the project, to build and run.
```
$ env GOOS=linux go build
$ ./asterisk-ari
```
