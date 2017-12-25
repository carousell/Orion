# handlers
`import "github.com/carousell/Orion/orion/handlers"`

* [Overview](#pkg-overview)
* [Imported Packages](#pkg-imports)
* [Index](#pkg-index)

## <a name="pkg-overview">Overview</a>

## <a name="pkg-imports">Imported Packages</a>

- [github.com/carousell/Orion/interceptors](./../../interceptors)
- [github.com/carousell/Orion/utils](./../../utils)
- [github.com/carousell/Orion/utils/headers](./../../utils/headers)
- [github.com/carousell/go-utils/utils/errors/notifier](./../../../go-utils/utils/errors/notifier)
- [github.com/gogo/protobuf/jsonpb](https://godoc.org/github.com/gogo/protobuf/jsonpb)
- [github.com/gogo/protobuf/proto](https://godoc.org/github.com/gogo/protobuf/proto)
- [github.com/gorilla/mux](https://godoc.org/github.com/gorilla/mux)
- [github.com/grpc-ecosystem/go-grpc-middleware/util/metautils](https://godoc.org/github.com/grpc-ecosystem/go-grpc-middleware/util/metautils)
- [github.com/grpc-ecosystem/go-grpc-prometheus](https://godoc.org/github.com/grpc-ecosystem/go-grpc-prometheus)
- [github.com/opentracing/opentracing-go](https://godoc.org/github.com/opentracing/opentracing-go)
- [google.golang.org/grpc](https://godoc.org/google.golang.org/grpc)
- [google.golang.org/grpc/codes](https://godoc.org/google.golang.org/grpc/codes)
- [google.golang.org/grpc/status](https://godoc.org/google.golang.org/grpc/status)

## <a name="pkg-index">Index</a>
* [Variables](#pkg-variables)
* [type Decodable](#Decodable)
* [type Decoder](#Decoder)
* [type Encodeable](#Encodeable)
* [type Encoder](#Encoder)
* [type GRPCMethodHandler](#GRPCMethodHandler)
* [type HTTPHandler](#HTTPHandler)
* [type HTTPHandlerConfig](#HTTPHandlerConfig)
* [type HTTPInterceptor](#HTTPInterceptor)
* [type Handler](#Handler)
  * [func NewGRPCHandler() Handler](#NewGRPCHandler)
  * [func NewHTTPHandler(config HTTPHandlerConfig) Handler](#NewHTTPHandler)
* [type Interceptor](#Interceptor)
* [type WhitelistedHeaders](#WhitelistedHeaders)

#### <a name="pkg-files">Package files</a>
[grpc.go](./grpc.go) [http.go](./http.go) [types.go](./types.go) [utils.go](./utils.go) 

## <a name="pkg-variables">Variables</a>
``` go
var (
    DefaultHTTPResponseHeaders = []string{
        "Content-Type",
    }
)
```

## <a name="Decodable">type</a> [Decodable](./types.go#L40-L42)
``` go
type Decodable interface {
    AddDecoder(serviceName, method string, decoder Decoder)
}
```
Decodable interface that is implemented by a handler that supports custom HTTP decoder

## <a name="Decoder">type</a> [Decoder](./types.go#L32)
``` go
type Decoder func(w http.ResponseWriter, decoderError, endpointError error, respObject interface{})
```

## <a name="Encodeable">type</a> [Encodeable](./types.go#L35-L37)
``` go
type Encodeable interface {
    AddEncoder(serviceName, method string, httpMethod []string, path string, encoder Encoder)
}
```
Encodeable interface that is implemented by a handler that supports custom HTTP encoder

## <a name="Encoder">type</a> [Encoder](./types.go#L30)
``` go
type Encoder func(req *http.Request, reqObject interface{}) error
```
Encoder is the function type needed for request encoders

## <a name="GRPCMethodHandler">type</a> [GRPCMethodHandler](./types.go#L13)
``` go
type GRPCMethodHandler func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error)
```
GRPCMethodHandler is the method type as defined in grpc-go

## <a name="HTTPHandler">type</a> [HTTPHandler](./types.go#L48)
``` go
type HTTPHandler func(http.ResponseWriter, *http.Request) bool
```

## <a name="HTTPHandlerConfig">type</a> [HTTPHandlerConfig](./http.go#L34-L36)
``` go
type HTTPHandlerConfig struct {
    EnableProtoURL bool
}
```
HTTPHandlerConfig is the configuration for HTTP Handler

## <a name="HTTPInterceptor">type</a> [HTTPInterceptor](./types.go#L44-L46)
``` go
type HTTPInterceptor interface {
    AddHTTPHandler(serviceName, method string, path string, handler HTTPHandler)
}
```

## <a name="Handler">type</a> [Handler](./types.go#L51-L55)
``` go
type Handler interface {
    Add(sd *grpc.ServiceDesc, ss interface{}) error
    Run(httpListener net.Listener) error
    Stop(timeout time.Duration) error
}
```
Handler implements a service handler that is used by orion server

### <a name="NewGRPCHandler">func</a> [NewGRPCHandler](./grpc.go#L14)
``` go
func NewGRPCHandler() Handler
```
NewGRPCHandler creates a new GRPC handler

### <a name="NewHTTPHandler">func</a> [NewHTTPHandler](./http.go#L39)
``` go
func NewHTTPHandler(config HTTPHandlerConfig) Handler
```
NewHTTPHandler creates a new HTTP handler

## <a name="Interceptor">type</a> [Interceptor](./types.go#L16-L19)
``` go
type Interceptor interface {
    // gets an array of Server Interceptors
    GetInterceptors() []grpc.UnaryServerInterceptor
}
```
Interceptor interface when implemented by a service allows that service to provide custom interceptors

## <a name="WhitelistedHeaders">type</a> [WhitelistedHeaders](./types.go#L22-L27)
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