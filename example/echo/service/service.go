// Package service must implement the generated proto's server interface
package service

import (
	"context"
	"strings"

	"github.com/carousell/Orion/example/echo/echo_proto"
	"github.com/carousell/Orion/interceptors"
	"google.golang.org/grpc"
)

type svc struct {
	appendText string
}

func GetService(config Config) echo_proto.EchoServiceServer {
	s := new(svc)
	s.appendText = config.AppendText
	return s
}

func (s *svc) Echo(ctx context.Context, req *echo_proto.EchoRequest) (*echo_proto.EchoResponse, error) {
	resp := new(echo_proto.EchoResponse)
	resp.Msg = s.appendText + req.GetMsg()
	return resp, nil
}

func (s *svc) Upper(ctx context.Context, req *echo_proto.UpperRequest) (*echo_proto.UpperResponse, error) {
	resp := new(echo_proto.UpperResponse)
	resp.Msg = strings.ToUpper(s.appendText + req.GetMsg())
	return resp, nil
}

func (s *svc) ABC(ctx context.Context, req *echo_proto.UpperRequest) (*echo_proto.UpperResponse, error) {
	resp := new(echo_proto.UpperResponse)
	resp.Msg = strings.ToUpper(s.appendText + req.GetMsg())
	return resp, nil
}

func (s *svc) GetInterceptors() []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{
		interceptors.DebugLoggingInterceptor(),
	}
}
