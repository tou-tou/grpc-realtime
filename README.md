# gRPC set up
[Quick start | Go | gRPC](https://grpc.io/docs/languages/go/quickstart/)

```
$ go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
$ go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
$ export PATH="$PATH:$(go env GOPATH)/bin"
```

# Regenerate gRPC code
While still in the `realtime-grpc`dir ( workspace root dir) , run the following command:
```
$ protoc --go_out=. --go_opt=paths=source_relative     --go-grpc_out=. --go-grpc_opt=paths=source_relative    world/user.proto world/room.proto
```
