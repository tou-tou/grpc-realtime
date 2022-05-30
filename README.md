# environment
```:bash
$ go version
go version go1.18.2 linux/amd64
$ protoc --version
libprotoc 3.12.4
```

# genererate gRPC code
In the `realtime-grpc`dir ( workspace root dir) , run the following command:
```
$ protoc --go_out=. --go_opt=paths=source_relative  --go-grpc_out=. --go-grpc_opt=paths=source_relative    proto/world/user.proto proto/world/room.proto 
```