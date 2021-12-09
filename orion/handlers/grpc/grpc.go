package grpc

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/carousell/Orion/orion/handlers"
	"github.com/carousell/Orion/utils/log"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
)

// Config is the configuration for GRPC Handler
type Config struct {
	handlers.CommonConfig
	CustomCodec           *grpc.Codec
	UnknownServiceHandler *grpc.StreamHandler
}

//NewGRPCHandler creates a new GRPC handler
func NewGRPCHandler(config Config) handlers.Handler {
	return &grpcHandler{config: config}
}

type grpcHandler struct {
	grpcServer  *grpc.Server
	mu          sync.Mutex
	config      Config
	middlewares *handlers.MiddlewareMapping
}

func (g *grpcHandler) init() {
	if g.grpcServer == nil {
		var opts []grpc.ServerOption
		if g.grpcInterceptor() != nil {
			opts = append(opts, grpc.UnaryInterceptor(g.grpcInterceptor()))
		}
		if g.grpcStreamInterceptor() != nil {
			opts = append(opts, grpc.StreamInterceptor(g.grpcStreamInterceptor()))
		}
		if g.config.CustomCodec != nil {
			opts = append(opts, grpc.CustomCodec(*g.config.CustomCodec))
		}
		if g.config.UnknownServiceHandler != nil {
			opts = append(opts, grpc.UnknownServiceHandler(*g.config.UnknownServiceHandler))
		}
		g.grpcServer = grpc.NewServer(opts...)
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
	log.Info(context.Background(), "GRPC", "server starting")
	grpc_prometheus.Register(g.grpcServer)
	return g.grpcServer.Serve(grpcListener)
}

func timedCall(f func(), timeout time.Duration) {
	done := make(chan struct{})
	go func() {
		f()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(timeout):
	}
}

func (g *grpcHandler) Stop(timeout time.Duration) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	log.Info(context.Background(), "GRPC", "stopping server")
	timedCall(g.grpcServer.GracefulStop, timeout)
	g.grpcServer.Stop()
	g.grpcServer = nil
	g.middlewares = nil
	log.Info(context.Background(), "GRPC", "stopped server")
	return nil
}

// grpcInterceptor acts as default interceptor for gprc and applies service specific interceptors based on implementation
func (g *grpcHandler) grpcInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// fetch method middlewares for this call
		middlewares := make([]string, 0)
		if g.middlewares != nil {
			middlewares = append(middlewares, g.middlewares.GetMiddlewaresFromURL(info.FullMethod)...)
		}
		// fetch interceptors from the service implementation and apply
		interceptor := handlers.GetInterceptorsWithMethodMiddlewares(info.Server, g.config.CommonConfig, middlewares)
		return interceptor(ctx, req, info, handler)
	}
}

// grpcStreamInterceptor acts as default interceptor for gprc streams and applies service specific interceptors based on implementation
func (g *grpcHandler) grpcStreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		interceptor := handlers.GetStreamInterceptors(srv, g.config.CommonConfig)
		return interceptor(srv, ss, info, handler)
	}
}
