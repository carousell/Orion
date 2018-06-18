# handlers
`import "github.com/carousell/Orion/orion/handlers"`

* [Overview](#pkg-overview)
* [Imported Packages](#pkg-imports)
* [Index](#pkg-index)

## <a name="pkg-overview">Overview</a>

## <a name="pkg-imports">Imported Packages</a>

- [github.com/carousell/Orion/interceptors](./../../interceptors)
- [github.com/carousell/Orion/orion/modifiers](./../modifiers)
- [github.com/carousell/Orion/utils/errors](./../../utils/errors)
- [github.com/carousell/Orion/utils/errors/notifier](./../../utils/errors/notifier)
- [github.com/carousell/Orion/utils/options](./../../utils/options)
- [google.golang.org/grpc](https://godoc.org/google.golang.org/grpc)

## <a name="pkg-index">Index</a>
* [func GetInterceptors(svc interface{}, config CommonConfig) grpc.UnaryServerInterceptor](#GetInterceptors)
* [func GetInterceptorsWithMethodMiddlewares(svc interface{}, config CommonConfig, middlewares []string) grpc.UnaryServerInterceptor](#GetInterceptorsWithMethodMiddlewares)
* [func GetMethodInterceptors(svc interface{}, config CommonConfig, middlewares []string) []grpc.UnaryServerInterceptor](#GetMethodInterceptors)
* [type CommonConfig](#CommonConfig)
* [type Decodable](#Decodable)
* [type Decoder](#Decoder)
* [type Encodeable](#Encodeable)
* [type Encoder](#Encoder)
* [type GRPCMethodHandler](#GRPCMethodHandler)
* [type HTTPHandler](#HTTPHandler)
* [type HTTPInterceptor](#HTTPInterceptor)
* [type Handler](#Handler)
* [type Interceptor](#Interceptor)
* [type MiddlewareMapping](#MiddlewareMapping)
  * [func NewMiddlewareMapping() \*MiddlewareMapping](#NewMiddlewareMapping)
  * [func (m \*MiddlewareMapping) AddMiddleware(service, method string, middlewares ...string)](#MiddlewareMapping.AddMiddleware)
  * [func (m \*MiddlewareMapping) GetMiddlewares(service, method string) []string](#MiddlewareMapping.GetMiddlewares)
  * [func (m \*MiddlewareMapping) GetMiddlewaresFromUrl(url string) []string](#MiddlewareMapping.GetMiddlewaresFromUrl)
* [type Middlewareable](#Middlewareable)
* [type Optionable](#Optionable)
* [type WhitelistedHeaders](#WhitelistedHeaders)

#### <a name="pkg-files">Package files</a>
[middleware.go](./middleware.go) [types.go](./types.go) [utils.go](./utils.go) 

## <a name="GetInterceptors">func</a> [GetInterceptors](./utils.go#L55)
``` go
func GetInterceptors(svc interface{}, config CommonConfig) grpc.UnaryServerInterceptor
```
GetInterceptors fetches interceptors from a given GRPC service

## <a name="GetInterceptorsWithMethodMiddlewares">func</a> [GetInterceptorsWithMethodMiddlewares](./utils.go#L59)
``` go
func GetInterceptorsWithMethodMiddlewares(svc interface{}, config CommonConfig, middlewares []string) grpc.UnaryServerInterceptor
```

## <a name="GetMethodInterceptors">func</a> [GetMethodInterceptors](./utils.go#L99)
``` go
func GetMethodInterceptors(svc interface{}, config CommonConfig, middlewares []string) []grpc.UnaryServerInterceptor
```

## <a name="CommonConfig">type</a> [CommonConfig](./types.go#L72-L74)
``` go
type CommonConfig struct {
    NoDefaultInterceptors bool
}
```
CommonConfig is the config that is common across both http and grpc handlers

## <a name="Decodable">type</a> [Decodable](./types.go#L42-L45)
``` go
type Decodable interface {
    AddDecoder(serviceName, method string, decoder Decoder)
    AddDefaultDecoder(serviceName string, decoder Decoder)
}
```
Decodable interface that is implemented by a handler that supports custom HTTP decoder

## <a name="Decoder">type</a> [Decoder](./types.go#L33)
``` go
type Decoder func(ctx context.Context, w http.ResponseWriter, encodeError, endpointError error, respObject interface{})
```
Decoder is the function type needed for response decoders

## <a name="Encodeable">type</a> [Encodeable](./types.go#L36-L39)
``` go
type Encodeable interface {
    AddEncoder(serviceName, method string, httpMethod []string, path string, encoder Encoder)
    AddDefaultEncoder(serviceName string, encoder Encoder)
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

## <a name="HTTPHandler">type</a> [HTTPHandler](./types.go#L57)
``` go
type HTTPHandler func(http.ResponseWriter, *http.Request) bool
```
HTTPHandler is the function that handles HTTP request

## <a name="HTTPInterceptor">type</a> [HTTPInterceptor](./types.go#L52-L54)
``` go
type HTTPInterceptor interface {
    AddHTTPHandler(serviceName, method string, path string, handler HTTPHandler)
}
```
HTTPInterceptor allows intercepting an HTTP connection

## <a name="Handler">type</a> [Handler](./types.go#L60-L64)
``` go
type Handler interface {
    Add(sd *grpc.ServiceDesc, ss interface{}) error
    Run(httpListener net.Listener) error
    Stop(timeout time.Duration) error
}
```
Handler implements a service handler that is used by orion server

## <a name="Interceptor">type</a> [Interceptor](./types.go#L16-L19)
``` go
type Interceptor interface {
    // gets an array of Server Interceptors
    GetInterceptors() []grpc.UnaryServerInterceptor
}
```
Interceptor interface when implemented by a service allows that service to provide custom interceptors

## <a name="MiddlewareMapping">type</a> [MiddlewareMapping](./middleware.go#L13-L15)
``` go
type MiddlewareMapping struct {
    // contains filtered or unexported fields
}
```

### <a name="NewMiddlewareMapping">func</a> [NewMiddlewareMapping](./middleware.go#L9)
``` go
func NewMiddlewareMapping() *MiddlewareMapping
```

### <a name="MiddlewareMapping.AddMiddleware">func</a> (\*MiddlewareMapping) [AddMiddleware](./middleware.go#L47)
``` go
func (m *MiddlewareMapping) AddMiddleware(service, method string, middlewares ...string)
```

### <a name="MiddlewareMapping.GetMiddlewares">func</a> (\*MiddlewareMapping) [GetMiddlewares](./middleware.go#L34)
``` go
func (m *MiddlewareMapping) GetMiddlewares(service, method string) []string
```

### <a name="MiddlewareMapping.GetMiddlewaresFromUrl">func</a> (\*MiddlewareMapping) [GetMiddlewaresFromUrl](./middleware.go#L29)
``` go
func (m *MiddlewareMapping) GetMiddlewaresFromUrl(url string) []string
```

## <a name="Middlewareable">type</a> [Middlewareable](./types.go#L67-L69)
``` go
type Middlewareable interface {
    AddMiddleware(serviceName, method string, middleware ...string)
}
```
Middlewareable implemets support for method specific middleware

## <a name="Optionable">type</a> [Optionable](./types.go#L47-L49)
``` go
type Optionable interface {
    AddOption(ServiceName, method, option string)
}
```

## <a name="WhitelistedHeaders">type</a> [WhitelistedHeaders](./types.go#L22-L27)
``` go
type WhitelistedHeaders interface {
    //GetRequestHeaders returns a list of all whitelisted request headers
    GetRequestHeaders() []string
    //GetResponseHeaders returns a list of all whitelisted response headers
    GetResponseHeaders() []string
}
```
WhitelistedHeaders is the interface that needs to be implemented by clients that need request/response headers to be passed in through the context

- - -
Generated by [godoc2ghmd](https://github.com/GandalfUK/godoc2ghmd)