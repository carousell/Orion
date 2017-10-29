package handlers

import (
	"context"
	"net"
	"net/http"

	"google.golang.org/grpc"
)

type RequestFunc func(ctx context.Context, request interface{}) context.Context
type ServerResponseFunc func(ctx context.Context, response interface{}) context.Context

type GRPCMethodHandler func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error)

type Interceptor interface {
	// gets an array of Server Interceptors
	GetInterceptors() []grpc.UnaryServerInterceptor
}

type Handler interface {
	Add(sd *grpc.ServiceDesc, ss interface{}) error
	Run(httpListener net.Listener, httpSrv *http.Server) error
}
