package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/carousell/Orion/example/ServiceName/ServiceName_proto"

	"google.golang.org/grpc"
)

const (
	address     = "192.168.99.100:9281"
	httpAddress = "http://192.168.99.100:9282/"
)

func main() {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := ServiceName_proto.NewServiceNameServiceClient(conn)
	echo(c)
	uppercase(c)
	echoHTTP()
	uppercaseHTTP()
	//
	addComment(c)
	searchComment(c)
}

func echo(c ServiceName_proto.ServiceNameServiceClient) {
	fmt.Println("making echo gRPC call")
	req := new(ServiceName_proto.EchoRequest)
	req.Msg = "Hello World"
	r, err := c.Echo(context.Background(), req)

	if err != nil {
		log.Fatalf("error: %v", err)
	}
	data, _ := json.Marshal(r)
	log.Printf("Response : %s", data)
}

func uppercase(c ServiceName_proto.ServiceNameServiceClient) {
	fmt.Println("making uppercase gRPC call")
	req := new(ServiceName_proto.UppercaseRequest)
	req.Msg = "Hello World"
	r, err := c.Uppercase(context.Background(), req)

	if err != nil {
		log.Fatalf("error: %v", err)
	}
	data, _ := json.Marshal(r)
	log.Printf("Response : %s", data)
}

func addComment(c ServiceName_proto.ServiceNameServiceClient) {
	fmt.Println("making addComment gRPC call")
	req := new(ServiceName_proto.AddCommentRequest)
	req.Comment = "Hello world test"
	r, err := c.AddComment(context.Background(), req)

	if err != nil {
		log.Fatalf("error: %v", err)
	}
	data, _ := json.Marshal(r)
	log.Printf("Response : %s", data)
	getComment(c, r.GetUUID())
}

func searchComment(c ServiceName_proto.ServiceNameServiceClient) {
	fmt.Println("making searchComment gRPC call")
	req := new(ServiceName_proto.SearchCommentsRequest)
	req.Query = "anot"
	r, err := c.SearchComments(context.Background(), req)

	if err != nil {
		log.Fatalf("error: %v", err)
	}
	data, _ := json.Marshal(r)
	log.Printf("Response : %s", data)
}

func getComment(c ServiceName_proto.ServiceNameServiceClient, uuid string) {
	fmt.Println("making getComment gRPC call with uuid ", uuid)
	req := new(ServiceName_proto.GetCommentRequest)
	req.UUID = uuid
	r, err := c.GetComment(context.Background(), req)

	if err != nil {
		log.Fatalf("error: %v", err)
	}
	data, _ := json.Marshal(r)
	log.Printf("Response : %s", data)
}

func echoHTTP() {
	fmt.Println("making echo HTTP call")
	r, err := http.Get(httpAddress + "echo/?msg=Test")
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	data, _ := ioutil.ReadAll(r.Body)
	log.Printf("Response : %s", data)
}

func uppercaseHTTP() {
	fmt.Println("making uppercase HTTP call")
	r, err := http.Get(httpAddress + "uppercase/?msg=Test")
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	data, _ := ioutil.ReadAll(r.Body)
	log.Printf("Response : %s", data)
}
