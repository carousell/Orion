package grpc

import (
	"context"
	"log"
	"net"
	"sync"
	"time"

	"github.com/carousell/Orion/orion/handlers"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
)

// GRPCConfig is the configuration for GRPC Handler
type GRPCConfig struct {
	handlers.CommonConfig
}

//NewGRPCHandler creates a new GRPC handler
func NewGRPCHandler(config GRPCConfig) handlers.Handler {
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
	log.Println("GRPC", "server starting")
	grpc_prometheus.Register(g.grpcServer)
	return g.grpcServer.Serve(grpcListener)
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

// grpcInterceptor acts as default interceptor for gprc and applies service specific interceptors based on implementation
func grpcInterceptor(config handlers.CommonConfig) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// fetch interceptors from the service implementation and apply
		interceptor := handlers.GetInterceptors(info.Server, config)
		return interceptor(ctx, req, info, handler)
	}
}
