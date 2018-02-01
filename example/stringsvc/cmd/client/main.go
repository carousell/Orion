package main

import (
	"context"
	"log"

	proto "github.com/carousell/Orion/example/stringsvc/stringproto"
	"google.golang.org/grpc"
)

const (
	address = "127.0.0.1:9281"
)

func main() {
	// establish connection
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := proto.NewStringServiceClient(conn)
	uppercase(client)
}

func uppercase(client proto.StringServiceClient) {
	r := new(proto.UpperRequest)
	r.Msg = "Hello World"
	log.Println("making gRPC calls for Upper")
	resp, err := client.Upper(context.Background(), r)
	log.Println("resp", resp)
	log.Println("error", err)
}
