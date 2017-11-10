package handlers

import (
	"context"
	"net"
	"net/http"
	"time"

	"google.golang.org/grpc"
)

type GRPCMethodHandler func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error)

type Interceptor interface {
	// gets an array of Server Interceptors
	GetInterceptors() []grpc.UnaryServerInterceptor
}

type Encoder func(req *http.Request, reqObject interface{}) error

type Encodeable interface {
	AddEncoder(serviceName, method, httpMethod, path string, encoder Encoder)
}

//Handler implements a service handler that is used by orion server
type Handler interface {
	Add(sd *grpc.ServiceDesc, ss interface{}) error
	Run(httpListener net.Listener) error
	Stop(timeout time.Duration) error
}
