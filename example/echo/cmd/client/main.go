package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	proto "github.com/carousell/Orion/example/echo/echo_proto"
	"google.golang.org/grpc"
)

const (
	//address = "192.168.99.100:9281"
	address = "127.0.0.1:9281"
)

func main() {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := proto.NewEchoServiceClient(conn)
	echo(c)
	uppercase(c)
}

func echo(c proto.EchoServiceClient) {
	fmt.Println("making echo gRPC call")
	req := new(proto.EchoRequest)
	req.Msg = "Hello World"
	r, err := c.Echo(context.Background(), req)

	if err != nil {
		log.Fatalf("error: %v", err)
	}
	data, _ := json.Marshal(r)
	log.Printf("Response : %s", data)
}

func uppercase(c proto.EchoServiceClient) {
	fmt.Println("making uppercase gRPC call")
	req := new(proto.UpperRequest)
	req.Msg = "Hello World"
	r, err := c.Upper(context.Background(), req)

	if err != nil {
		log.Fatalf("error: %v", err)
	}
	data, _ := json.Marshal(r)
	log.Printf("Response : %s", data)
}
