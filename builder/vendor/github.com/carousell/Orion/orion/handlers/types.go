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
	//GetRequestHeaders returns a list of all whitelisted request headers
	GetRequestHeaders() []string
	//GetResponseHeaders returns a list of all whitelisted response headers
	GetResponseHeaders() []string
}

//Encoder is the function type needed for request encoders
type Encoder func(req *http.Request, reqObject interface{}) error

//Decoder is the function type needed for response decoders
type Decoder func(ctx context.Context, w http.ResponseWriter, encodeError, endpointError error, respObject interface{})

//Encodeable interface that is implemented by a handler that supports custom HTTP encoder
type Encodeable interface {
	AddEncoder(serviceName, method string, httpMethod []string, path string, encoder Encoder)
	AddDefaultEncoder(serviceName string, encoder Encoder)
}

//Decodable interface that is implemented by a handler that supports custom HTTP decoder
type Decodable interface {
	AddDecoder(serviceName, method string, decoder Decoder)
	AddDefaultDecoder(serviceName string, decoder Decoder)
}

type Optionable interface {
	AddOption(ServiceName, method, option string)
}

//HTTPInterceptor allows intercepting an HTTP connection
type HTTPInterceptor interface {
	AddHTTPHandler(serviceName, method string, path string, handler HTTPHandler)
}

// HTTPHandler is the function that handles HTTP request
type HTTPHandler func(http.ResponseWriter, *http.Request) bool

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
	NoDefaultInterceptors bool
}
