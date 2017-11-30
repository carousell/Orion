// Package service must implement the generated proto's server interface
package service

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	proto "github.com/carousell/Orion/example/echo/echo_proto"
	"github.com/carousell/Orion/interceptors"
	"github.com/carousell/Orion/utils/headers"
	"github.com/carousell/Orion/utils/spanutils"
	"google.golang.org/grpc"
)

const (
	//address = "192.168.99.100:9281"
	address = "127.0.0.1:9281"
)

type svc struct {
	appendText string
	debug      bool
	client     proto.EchoServiceClient
}

func (s *svc) GetRequestHeaders() []string {
	return []string{}
}

func (s *svc) GetResponseHeaders() []string {
	return []string{"Original-Msg"}
}

func GetService(config Config) proto.EchoServiceServer {
	s := new(svc)
	s.appendText = config.AppendText
	s.debug = config.Debug
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithUnaryInterceptor(interceptors.GRPCClientInterceptor()))
	if err != nil {
		log.Fatalln("did not connect: %v", err)
	}
	s.client = proto.NewEchoServiceClient(conn)
	return s
}

func (s *svc) Echo(ctx context.Context, req *proto.EchoRequest) (*proto.EchoResponse, error) {
	resp := new(proto.EchoResponse)
	resp.Msg = s.appendText + req.GetMsg()

	httpClient := &http.Client{}
	url := "http://127.0.0.1:9282/api/1.0/upper/" + req.GetMsg()
	httpReq, _ := http.NewRequest("GET", url, nil)
	sp, ctx := spanutils.NewHTTPExternalSpan(ctx, "helloworld", url, httpReq.Header)
	defer sp.Finish()
	//log.Println(httpReq)
	httpClient.Do(httpReq)

	r := new(proto.UpperRequest)
	r.Msg = "hello"
	s.client.Upper(ctx, r)
	return resp, nil
}

func (s *svc) Upper(ctx context.Context, req *proto.UpperRequest) (*proto.UpperResponse, error) {
	resp := new(proto.UpperResponse)
	resp.Msg = strings.ToUpper(s.appendText + req.GetMsg())
	/*
		hdrs := headers.RequestHeadersFromContext(ctx)
		if hdrs != nil {
			fmt.Println("All request headers", hdrs.GetAll())
		}
	*/
	headers.AddToResponseHeaders(ctx, "original-msg", req.GetMsg())
	sp, _ := spanutils.NewDatastoreSpan(ctx, "Wait", "Wa")
	defer sp.End()
	time.Sleep(100 * time.Millisecond)
	return resp, nil
}

func (s *svc) GetInterceptors() []grpc.UnaryServerInterceptor {
	icpt := []grpc.UnaryServerInterceptor{}
	if s.debug {
		icpt = append(icpt, interceptors.DebugLoggingInterceptor())
	}
	return icpt
}

func (s *svc) UpperProxy(ctx context.Context, req *proto.UpperRequest) (*proto.UpperResponse, error) {
	return s.client.Upper(ctx, req)
}
