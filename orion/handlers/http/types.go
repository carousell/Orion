package http

import (
	"net/http"
	"sync"

	"github.com/carousell/Orion/orion/handlers"
	"github.com/carousell/Orion/orion/modifiers"
	"github.com/gogo/protobuf/jsonpb"
	"google.golang.org/grpc"
)

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

	// DefaultHTTPResponseHeaders are response headers that are whitelisted by default
	DefaultHTTPResponseHeaders = []string{
		"Content-Type",
	}
)

//HTTPHandlerConfig is the configuration for HTTP Handler
type HTTPHandlerConfig struct {
	handlers.CommonConfig
	EnableProtoURL bool
}

type serviceInfo struct {
	desc            *grpc.ServiceDesc
	svc             interface{}
	interceptors    grpc.UnaryServerInterceptor
	requestHeaders  []string
	responseHeaders []string
}

type methodInfo struct {
	svc         *serviceInfo
	method      handlers.GRPCMethodHandler
	encoder     handlers.Encoder
	decoder     handlers.Decoder
	httpHandler handlers.HTTPHandler
	httpMethod  []string
	encoderPath string
	serviceName string
	methodName  string
	urls        []string
}

type httpHandler struct {
	mu          sync.Mutex
	mapping     *methodInfoMapping
	defEncoders map[string]handlers.Encoder
	defDecoders map[string]handlers.Decoder
	mar         jsonpb.Marshaler
	svr         *http.Server
	config      HTTPHandlerConfig
}
