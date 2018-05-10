package handlers

import (
	"log"
	"net"
	"sync"
	"time"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
)

// GRPCConfig is the configuration for GRPC Handler
type GRPCConfig struct {
	CommonConfig
}

//NewGRPCHandler creates a new GRPC handler
func NewGRPCHandler(config GRPCConfig) Handler {
	return &grpcHandler{config: config}
}

type grpcHandler struct {
	grpcServer *grpc.Server
	mu         sync.Mutex
	config     GRPCConfig
}

func (g *grpcHandler) init() {
	if g.grpcServer == nil {
		opt := make([]grpc.ServerOption, 0)
		opt = append(opt, grpc.UnaryInterceptor(grpcInterceptor(g.config.CommonConfig)))
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
	g.mu.Lock()
	log.Println("GRPC", "server starting")
	grpc_prometheus.Register(g.grpcServer)
	s := g.grpcServer
	g.mu.Unlock()
	return s.Serve(grpcListener)
}

func (g *grpcHandler) Stop(timeout time.Duration) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	log.Println("GRPC", "stopping server")
	s := g.grpcServer
	g.grpcServer = nil
	s.GracefulStop()
	time.Sleep(timeout)
	s.Stop()
	log.Println("GRPC", "stopped server")
	return nil
}
