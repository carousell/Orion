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

func (g *grpcHandler) init() {
	if g.grpcServer == nil {
		opt := make([]grpc.ServerOption, 0)
		opt = append(opt, grpc.UnaryInterceptor(grpcInterceptor()))
		g.grpcServer = grpc.NewServer(opt...)
	}
}

func (g *grpcHandler) Add(sd *grpc.ServiceDesc, ss interface{}) error {
	g.init()
	g.grpcServer.RegisterService(sd, ss)
	return nil
}

func (g *grpcHandler) Run(grpcListener net.Listener) error {
	return g.grpcServer.Serve(grpcListener)
}

func (g *grpcHandler) Stop(timeout time.Duration) error {
	g.grpcServer.GracefulStop()
	go func(s *grpc.Server) {
		time.AfterFunc(timeout, func() {
			s.Stop()
		})
	}(g.grpcServer)
	g.grpcServer = nil
	return nil
}
