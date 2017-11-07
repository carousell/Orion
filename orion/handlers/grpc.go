package handlers

import (
	"log"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
)

//NewGRPCHandler creates a new GRPC handler
func NewGRPCHandler() Handler {
	return &grpcHandler{}
}

type grpcHandler struct {
	grpcServer *grpc.Server
	mu         sync.Mutex
}

func (g *grpcHandler) init() {
	if g.grpcServer == nil {
		opt := make([]grpc.ServerOption, 0)
		opt = append(opt, grpc.UnaryInterceptor(grpcInterceptor()))
		g.grpcServer = grpc.NewServer(opt...)
	}
}

func (g *grpcHandler) Add(sd *grpc.ServiceDesc, ss interface{}) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.init()
	g.grpcServer.RegisterService(sd, ss)
	return nil
}

func (g *grpcHandler) Run(grpcListener net.Listener) error {
	log.Println("GRPC", "server starting")
	return g.grpcServer.Serve(grpcListener)
}

func (g *grpcHandler) Stop(timeout time.Duration) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	log.Println("GRPC", "stopping server")
	s := g.grpcServer
	g.grpcServer = nil
	time.Sleep(timeout)
	s.Stop()
	log.Println("GRPC", "stopped server")
	return nil
}
