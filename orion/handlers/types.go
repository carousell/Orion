package handlers

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/carousell/Orion/orion/modifiers"
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

//Decoder is the function type needed for response decoders
type Decoder func(ctx context.Context, w http.ResponseWriter, decoderError, endpointError error, respObject interface{})

//Encodeable interface that is implemented by a handler that supports custom HTTP encoder
type Encodeable interface {
	AddEncoder(serviceName, method string, httpMethod []string, path string, encoder Encoder)
}

//Decodable interface that is implemented by a handler that supports custom HTTP decoder
type Decodable interface {
	AddDecoder(serviceName, method string, decoder Decoder)
}

//HTTPInterceptor allows intercepting an HTTP connection
type HTTPInterceptor interface {
	AddHTTPHandler(serviceName, method string, path string, handler HTTPHandler)
}

// HTTPHandler is the funtion that handles HTTP request
type HTTPHandler func(http.ResponseWriter, *http.Request) bool

//Handler implements a service handler that is used by orion server
type Handler interface {
	Add(sd *grpc.ServiceDesc, ss interface{}) error
	Run(httpListener net.Listener) error
	Stop(timeout time.Duration) error
}

//CommonConfig is the config that is common across both http and grpc handlers
type CommonConfig struct {
	NoDefaultInterceptors bool
}

var (
	//ContentTypeMap is the mapping of content-type with marshaling type
	ContentTypeMap = map[string]string{
		"application/json":                modifiers.JSON,
		"application/jsonpb":              modifiers.JSONPB,
		"application/x-jsonpb":            modifiers.JSONPB,
		"application/protobuf":            modifiers.ProtoBuf,
		"application/proto":               modifiers.ProtoBuf,
		"application/x-proto":             modifiers.ProtoBuf,
		"application/vnd.google.protobuf": modifiers.ProtoBuf,
		"application/octet-stream":        modifiers.ProtoBuf,
	}
)
