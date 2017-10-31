package handlers

import (
	"context"
	"net"
	"time"

	"google.golang.org/grpc"
)

type GRPCMethodHandler func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error)

type Interceptor interface {
	// gets an array of Server Interceptors
	GetInterceptors() []grpc.UnaryServerInterceptor
}

type Handler interface {
	Add(sd *grpc.ServiceDesc, ss interface{}) error
	Run(httpListener net.Listener) error
	Stop(timeout time.Duration) error
}
