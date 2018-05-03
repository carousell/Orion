# handlers
`import "github.com/carousell/Orion/orion/handlers"`

* [Overview](#pkg-overview)
* [Imported Packages](#pkg-imports)
* [Index](#pkg-index)

## <a name="pkg-overview">Overview</a>

## <a name="pkg-imports">Imported Packages</a>

- [github.com/carousell/Orion/interceptors](./../../interceptors)
- [github.com/carousell/Orion/orion/modifiers](./../modifiers)
- [github.com/carousell/Orion/utils](./../../utils)
- [github.com/carousell/Orion/utils/errors/notifier](./../../utils/errors/notifier)
- [github.com/carousell/Orion/utils/headers](./../../utils/headers)
- [github.com/carousell/Orion/utils/options](./../../utils/options)
- [github.com/golang/protobuf/jsonpb](https://godoc.org/github.com/golang/protobuf/jsonpb)
- [github.com/golang/protobuf/proto](https://godoc.org/github.com/golang/protobuf/proto)
- [github.com/gorilla/mux](https://godoc.org/github.com/gorilla/mux)
- [github.com/grpc-ecosystem/go-grpc-middleware/util/metautils](https://godoc.org/github.com/grpc-ecosystem/go-grpc-middleware/util/metautils)
- [github.com/grpc-ecosystem/go-grpc-prometheus](https://godoc.org/github.com/grpc-ecosystem/go-grpc-prometheus)
- [github.com/mitchellh/mapstructure](https://godoc.org/github.com/mitchellh/mapstructure)
- [github.com/opentracing/opentracing-go](https://godoc.org/github.com/opentracing/opentracing-go)
- [google.golang.org/grpc](https://godoc.org/google.golang.org/grpc)
- [google.golang.org/grpc/codes](https://godoc.org/google.golang.org/grpc/codes)
- [google.golang.org/grpc/status](https://godoc.org/google.golang.org/grpc/status)

## <a name="pkg-index">Index</a>
* [Variables](#pkg-variables)
* [func ContentTypeFromHeaders(ctx context.Context) string](#ContentTypeFromHeaders)
* [func DefaultEncoder(req \*http.Request, r interface{}) error](#DefaultEncoder)
* [func GrpcErrorToHTTP(err error, defaultStatus int, defaultMessage string) (int, string)](#GrpcErrorToHTTP)
* [type CommonConfig](#CommonConfig)
* [type Decodable](#Decodable)
* [type Decoder](#Decoder)
* [type Encodeable](#Encodeable)
* [type Encoder](#Encoder)
* [type GRPCConfig](#GRPCConfig)
* [type GRPCMethodHandler](#GRPCMethodHandler)
* [type HTTPHandler](#HTTPHandler)
* [type HTTPHandlerConfig](#HTTPHandlerConfig)
* [type HTTPInterceptor](#HTTPInterceptor)
* [type Handler](#Handler)
  * [func NewGRPCHandler(config GRPCConfig) Handler](#NewGRPCHandler)
  * [func NewHTTPHandler(config HTTPHandlerConfig) Handler](#NewHTTPHandler)
* [type Interceptor](#Interceptor)
* [type WhitelistedHeaders](#WhitelistedHeaders)

#### <a name="pkg-files">Package files</a>
[grpc.go](./grpc.go) [http.go](./http.go) [types.go](./types.go) [utils.go](./utils.go) 

## <a name="pkg-variables">Variables</a>
``` go
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
```
``` go
var (
    // DefaultHTTPResponseHeaders are reponse headers that are whitelisted by default
    DefaultHTTPResponseHeaders = []string{
        "Content-Type",
    }
)
```

## <a name="ContentTypeFromHeaders">func</a> [ContentTypeFromHeaders](./utils.go#L110)
``` go
func ContentTypeFromHeaders(ctx context.Context) string
```
ContentTypeFromHeaders searches for a matching content type

## <a name="DefaultEncoder">func</a> [DefaultEncoder](./http.go#L228)
``` go
func DefaultEncoder(req *http.Request, r interface{}) error
```
DefaultEncoder encodes a HTTP request if none are registered. This encoder
populates the proto message with URL route variables or fields from a JSON
body if either are available.

## <a name="GrpcErrorToHTTP">func</a> [GrpcErrorToHTTP](./http.go#L180)
``` go
func GrpcErrorToHTTP(err error, defaultStatus int, defaultMessage string) (int, string)
```
GrpcErrorToHTTP converts gRPC error code into HTTP response status code.
See: <a href="https://github.com/googleapis/googleapis/blob/master/google/rpc/code.proto">https://github.com/googleapis/googleapis/blob/master/google/rpc/code.proto</a>

## <a name="CommonConfig">type</a> [CommonConfig](./types.go#L62-L64)
``` go
type CommonConfig struct {
    NoDefaultInterceptors bool
}
```
CommonConfig is the config that is common across both http and grpc handlers

## <a name="Decodable">type</a> [Decodable](./types.go#L42-L44)
``` go
type Decodable interface {
    AddDecoder(serviceName, method string, decoder Decoder)
}
```
Decodable interface that is implemented by a handler that supports custom HTTP decoder

## <a name="Decoder">type</a> [Decoder](./types.go#L34)
``` go
type Decoder func(ctx context.Context, w http.ResponseWriter, encodeError, endpointError error, respObject interface{})
```
Decoder is the function type needed for response decoders

## <a name="Encodeable">type</a> [Encodeable](./types.go#L37-L39)
``` go
type Encodeable interface {
    AddEncoder(serviceName, method string, httpMethod []string, path string, encoder Encoder)
}
```
Encodeable interface that is implemented by a handler that supports custom HTTP encoder

## <a name="Encoder">type</a> [Encoder](./types.go#L31)
``` go
type Encoder func(req *http.Request, reqObject interface{}) error
```
Encoder is the function type needed for request encoders

## <a name="GRPCConfig">type</a> [GRPCConfig](./grpc.go#L14-L16)
``` go
type GRPCConfig struct {
    CommonConfig
}
```
GRPCConfig is the configuration for GRPC Handler

## <a name="GRPCMethodHandler">type</a> [GRPCMethodHandler](./types.go#L14)
``` go
type GRPCMethodHandler func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error)
```
GRPCMethodHandler is the method type as defined in grpc-go

## <a name="HTTPHandler">type</a> [HTTPHandler](./types.go#L52)
``` go
type HTTPHandler func(http.ResponseWriter, *http.Request) bool
```
HTTPHandler is the funtion that handles HTTP request

## <a name="HTTPHandlerConfig">type</a> [HTTPHandlerConfig](./http.go#L40-L43)
``` go
type HTTPHandlerConfig struct {
    CommonConfig
    EnableProtoURL bool
}
```
HTTPHandlerConfig is the configuration for HTTP Handler

## <a name="HTTPInterceptor">type</a> [HTTPInterceptor](./types.go#L47-L49)
``` go
type HTTPInterceptor interface {
    AddHTTPHandler(serviceName, method string, path string, handler HTTPHandler)
}
```
HTTPInterceptor allows intercepting an HTTP connection

## <a name="Handler">type</a> [Handler](./types.go#L55-L59)
``` go
type Handler interface {
    Add(sd *grpc.ServiceDesc, ss interface{}) error
    Run(httpListener net.Listener) error
    Stop(timeout time.Duration) error
}
```
Handler implements a service handler that is used by orion server

### <a name="NewGRPCHandler">func</a> [NewGRPCHandler](./grpc.go#L19)
``` go
func NewGRPCHandler(config GRPCConfig) Handler
```
NewGRPCHandler creates a new GRPC handler

### <a name="NewHTTPHandler">func</a> [NewHTTPHandler](./http.go#L46)
``` go
func NewHTTPHandler(config HTTPHandlerConfig) Handler
```
NewHTTPHandler creates a new HTTP handler

## <a name="Interceptor">type</a> [Interceptor](./types.go#L17-L20)
``` go
type Interceptor interface {
    // gets an array of Server Interceptors
    GetInterceptors() []grpc.UnaryServerInterceptor
}
```
Interceptor interface when implemented by a service allows that service to provide custom interceptors

## <a name="WhitelistedHeaders">type</a> [WhitelistedHeaders](./types.go#L23-L28)
``` go
type WhitelistedHeaders interface {
    //GetRequestHeaders retuns a list of all whitelisted request headers
    GetRequestHeaders() []string
    //GetResponseHeaders retuns a list of all whitelisted response headers
    GetResponseHeaders() []string
}
```
WhitelistedHeaders is the interface that needs to be implemented by clients that need request/response headers to be passed in through the context

- - -
Generated by [godoc2ghmd](https://github.com/GandalfUK/godoc2ghmd)