package grpc

import (
	"context"
	"fmt"
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
	grpcServer  *grpc.Server
	mu          sync.Mutex
	config      GRPCConfig
	middlewares *handlers.MiddlewareMapping
}

func (g *grpcHandler) init() {
	if g.grpcServer == nil {
		opt := make([]grpc.ServerOption, 0)
		opt = append(opt, grpc.UnaryInterceptor(g.grpcInterceptor()))
		g.grpcServer = grpc.NewServer(opt...)
	}
	if g.middlewares == nil {
		g.middlewares = handlers.NewMiddlewareMapping()
	}
}

func (g *grpcHandler) Add(sd *grpc.ServiceDesc, ss interface{}) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.init()
	g.grpcServer.RegisterService(sd, ss)
	return nil
}

func (g *grpcHandler) AddMiddleware(serviceName string, method string, middlewares ...string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.init()
	g.middlewares.AddMiddleware(serviceName, method, middlewares...)
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
func (g *grpcHandler) grpcInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// fetch method middlewares for this call
		middlewares := make([]string, 0)
		if g.middlewares != nil {
			middlewares = append(middlewares, g.middlewares.GetMiddlewaresFromUrl(info.FullMethod)...)
		}
		fmt.Println("GRPC middlewares", middlewares)
		// fetch interceptors from the service implementation and apply
		interceptor := handlers.GetInterceptorsWithMethodMiddlewares(info.Server, g.config.CommonConfig, middlewares)
		return interceptor(ctx, req, info, handler)
	}
}
