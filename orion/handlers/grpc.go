package handlers

import (
	"net"
	"time"

	"google.golang.org/grpc"
)

//NewGRPCHandler creates a new GRPC handler
func NewGRPCHandler() Handler {
	return &grpcHandler{}
}

type grpcHandler struct {
	grpcServer *grpc.Server
}

func (g *grpcHandler) Add(sd *grpc.ServiceDesc, ss interface{}) error {
	opt := make([]grpc.ServerOption, 0)

	opt = append(opt, grpc.UnaryInterceptor(getInterceptors(ss)))

	g.grpcServer = grpc.NewServer(opt...)
	g.grpcServer.RegisterService(sd, ss)
	return nil
}

func (g *grpcHandler) Run(grpcListener net.Listener) error {
	return g.grpcServer.Serve(grpcListener)
}

func (g *grpcHandler) Stop(timeout time.Duration) error {
	return nil
}
