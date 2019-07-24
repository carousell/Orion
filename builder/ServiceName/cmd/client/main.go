package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	proto "github.com/carousell/Orion/builder/ServiceName/ServiceName_proto"
	"github.com/carousell/Orion/utils/errors"
	"google.golang.org/grpc"
)

const (
	address = "192.168.99.100:9281"
	// address = "127.0.0.1:9281"
)

func main() {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := proto.NewServiceNameClient(conn)
	// echo(c)
	// uppercase(c)
	testStreamInterceptor(c)
}

func echo(c proto.ServiceNameClient) {
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

func uppercase(c proto.ServiceNameClient) {
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

func testStreamInterceptor(c proto.ServiceNameClient) {
	stream, err := c.TestStreamInterceptor(context.Background())
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to create stream"))
	}
	defer stream.CloseSend()
	reqs := []*proto.TestStreamInterceptorRequest{
		&proto.TestStreamInterceptorRequest{SleepMs: 50},
		&proto.TestStreamInterceptorRequest{SleepMs: 50},
		&proto.TestStreamInterceptorRequest{SleepMs: 50},
		// &proto.TestStreamInterceptorRequest{ShouldReturnError: true}, // will return error
	}
	for _, req := range reqs {
		if err := stream.Send(req); err != nil {
			log.Fatal(err)
		}
	}
	resp, err := stream.CloseAndRecv()
	fmt.Printf("resp = %+v\n", resp)
	fmt.Printf("err = %+v\n", err)
}
