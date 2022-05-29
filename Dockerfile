FROM golang:1.18

RUN apt-get update  \
    && apt-get install unzip \
    && apt-get install -y protobuf-compiler


WORKDIR /go/src/github.com/tou-tou/realtim-grpc


#COPY go.mod go.sum ./
#RUN go mod download
#EXPOSE 8080

#CMD ["go", "run", "main.go"]