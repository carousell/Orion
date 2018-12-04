package http

import (
	"net/http"
	"sync"

	"github.com/carousell/Orion/orion/handlers"
	"github.com/carousell/Orion/orion/modifiers"
	"github.com/golang/protobuf/jsonpb"
	"google.golang.org/grpc"
)

var (
	//ContentTypeMap is the mapping of content-type with marshaling type
	ContentTypeMap = map[string]string{
		ContentTypeJSON:                   modifiers.JSON,
		"application/jsonpb":              modifiers.JSONPB,
		"application/x-jsonpb":            modifiers.JSONPB,
		"application/protobuf":            modifiers.ProtoBuf,
		"application/proto":               modifiers.ProtoBuf,
		"application/x-proto":             modifiers.ProtoBuf,
		"application/vnd.google.protobuf": modifiers.ProtoBuf,
		ContentTypeProto:                  modifiers.ProtoBuf,
	}

	// DefaultHTTPResponseHeaders are response headers that are whitelisted by default
	DefaultHTTPResponseHeaders = []string{
		"Content-Type",
	}
)

const (
	//IgnoreNR is the option flag to ignore newrelic for this method
	IgnoreNR = "IGNORE_NR"
)

const (
	ContentTypeJSON  = "application/json"
	ContentTypeProto = "application/octet-stream"
)

//Config is the configuration for HTTP Handler
type Config struct {
	handlers.CommonConfig
	EnableProtoURL bool
}

type serviceInfo struct {
	desc            *grpc.ServiceDesc
	svc             interface{}
	requestHeaders  []string
	responseHeaders []string
}

type methodInfo struct {
	svc           *serviceInfo
	method        handlers.GRPCMethodHandler
	stream        grpc.StreamHandler
	encoder       handlers.Encoder
	decoder       handlers.Decoder
	httpHandler   handlers.HTTPHandler
	httpMethod    []string
	encoderPath   string
	serviceName   string
	methodName    string
	urls          []string
	options       []string
	clientStreams bool
	serverStreams bool
}

type httpHandler struct {
	mu          sync.Mutex
	mapping     *methodInfoMapping
	middlewares *handlers.MiddlewareMapping
	defEncoders map[string]handlers.Encoder
	defDecoders map[string]handlers.Decoder
	mar         jsonpb.Marshaler
	svr         *http.Server
	config      Config
}
