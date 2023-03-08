package handlers

import (
	"context"
	"net"
	"time"

	"google.golang.org/grpc"
)

//GRPCMethodHandler is the method type as defined in grpc-go
type GRPCMethodHandler func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error)

//Interceptor interface when implemented by a service allows that service to provide custom interceptors
type Interceptor interface {
	// gets an array of Unary Server Interceptors
	GetInterceptors() []grpc.UnaryServerInterceptor
}

//StreamInterceptor interface when implemented by a service allows that service to provide custom stream interceptors
type StreamInterceptor interface {
	// gets an array of Stream Server Interceptors
	GetStreamInterceptors() []grpc.StreamServerInterceptor
}

//WhitelistedHeaders is the interface that needs to be implemented by clients that need request/response headers to be passed in through the context
type WhitelistedHeaders interface {
	//GetRequestHeaders returns a list of all whitelisted request headers
	GetRequestHeaders() []string
	//GetResponseHeaders returns a list of all whitelisted response headers
	GetResponseHeaders() []string
}

//Optionable interface that is implemented by a handler that support custom Orion options
type Optionable interface {
	AddOption(ServiceName, method, option string)
}

//Handler implements a service handler that is used by orion server
type Handler interface {
	Add(sd *grpc.ServiceDesc, ss interface{}) error
	Run(httpListener net.Listener) error
	Stop(timeout time.Duration) error
}

//Middlewareable implemets support for method specific middleware
type Middlewareable interface {
	AddMiddleware(serviceName, method string, middleware ...string)
}

//CommonConfig is the config that is common across both http and grpc handlers
type CommonConfig struct {
	DisableDefaultInterceptors bool
}
