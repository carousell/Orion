package service

import (
	"context"
	"log"
	"strings"

	proto "github.com/carousell/Orion/example/stringsvc2/stringproto"
	"google.golang.org/grpc"
)

func NewSvc(config Config) proto.StringServiceServer {
	return &svc{
		debug: config.Debug,
	}
}

// svc implements proto.StringServiceServer
type svc struct {
	debug bool
}

func (s *svc) Upper(ctx context.Context, req *proto.UpperRequest) (*proto.UpperResponse, error) {
	resp := new(proto.UpperResponse)
	resp.Msg = strings.ToUpper(req.GetMsg())
	return resp, nil
}

func (s *svc) Count(ctx context.Context, req *proto.CountRequest) (*proto.CountResponse, error) {
	resp := new(proto.CountResponse)
	resp.Count = int64(len(req.GetMsg()))
	return resp, nil
}

func (s *svc) GetInterceptors() []grpc.UnaryServerInterceptor {
	if s.debug {
		return []grpc.UnaryServerInterceptor{DebugInterceptor}
	} else {
		return []grpc.UnaryServerInterceptor{}
	}
}

func DebugInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	log.Println("Method", info.FullMethod)
	log.Println("Request", req)
	resp, err = handler(ctx, req)
	log.Println("Response", resp)
	log.Println("Error", err)
	return
}
