package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	proto "github.com/carousell/Orion/example/simple/simple_proto"
	"google.golang.org/grpc"
)

const (
	//address = "192.168.99.100:9281"
	address  = "127.0.0.1"
	grpcPort = "9281"
	httpPort = "9282"
)

func main() {
	conn, err := grpc.Dial(address+":"+grpcPort, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := proto.NewSimpleServiceClient(conn)
	echoGRPC(c)
}

func echoGRPC(c proto.SimpleServiceClient) {
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

func echoHTTP() {
}
