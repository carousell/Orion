// Package service must implement the generated proto's server interface
package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/carousell/Orion/example/echo/echo_proto"
	"google.golang.org/grpc"
)

type svc struct{}

func GetService() echo_proto.EchoServiceServer {
	return new(svc)
}

func (s *svc) Echo(ctx context.Context, req *echo_proto.EchoRequest) (*echo_proto.EchoResponse, error) {
	resp := new(echo_proto.EchoResponse)
	resp.Msg = req.GetMsg()
	return resp, nil
}

func (s *svc) Upper(ctx context.Context, req *echo_proto.UpperRequest) (*echo_proto.UpperResponse, error) {
	resp := new(echo_proto.UpperResponse)
	resp.Msg = strings.ToUpper(req.GetMsg())
	return resp, nil
}

func (s *svc) ABC(ctx context.Context, req *echo_proto.UpperRequest) (*echo_proto.UpperResponse, error) {
	resp := new(echo_proto.UpperResponse)
	resp.Msg = strings.ToUpper(req.GetMsg())
	return resp, nil
}

func (s *svc) GetInterceptors() []grpc.UnaryServerInterceptor {

	return []grpc.UnaryServerInterceptor{
		func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			fmt.Println(info, "requst", req)
			resp, err := handler(ctx, req)
			fmt.Println(info, "response", resp, err)
			return resp, err
		},
	}
}
