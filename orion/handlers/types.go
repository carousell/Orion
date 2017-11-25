package handlers

import (
	"context"
	"net"
	"net/http"
	"time"

	"google.golang.org/grpc"
)

//GRPCMethodHandler is the method type as defined in grpc-go
type GRPCMethodHandler func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error)

//Interceptor interface when implemented by a service allows that service to provide custom interceptors
type Interceptor interface {
	// gets an array of Server Interceptors
	GetInterceptors() []grpc.UnaryServerInterceptor
}

//WhitelistedHeaders is the interface that needs to be implemented by clients that need request/response headers to be passed in through the context
type WhitelistedHeaders interface {
	//GetRequestHeaders retuns a list of all whitelisted request headers
	GetRequestHeaders() []string
	//GetResponseHeaders retuns a list of all whitelisted response headers
	GetResponseHeaders() []string
}

//Encoder is the function type needed for request encoders
type Encoder func(req *http.Request, reqObject interface{}) error

//Encodeable interface that is implemented by a handler that supports custom HTTP encoder
type Encodeable interface {
	AddEncoder(serviceName, method, httpMethod, path string, encoder Encoder)
}

//Handler implements a service handler that is used by orion server
type Handler interface {
	Add(sd *grpc.ServiceDesc, ss interface{}) error
	Run(httpListener net.Listener) error
	Stop(timeout time.Duration) error
}
