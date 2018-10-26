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
- [github.com/carousell/Orion/utils/log](./../../utils/log)
- [github.com/carousell/Orion/utils/log/loggers](./../../utils/log/loggers)
- [github.com/carousell/Orion/utils/options](./../../utils/options)
- [google.golang.org/grpc](https://godoc.org/google.golang.org/grpc)

## <a name="pkg-index">Index</a>
* [func GetInterceptors(svc interface{}, config CommonConfig) grpc.UnaryServerInterceptor](#GetInterceptors)
* [func GetInterceptorsWithMethodMiddlewares(svc interface{}, config CommonConfig, middlewares []string) grpc.UnaryServerInterceptor](#GetInterceptorsWithMethodMiddlewares)
* [func GetMethodInterceptors(svc interface{}, config CommonConfig, middlewares []string) []grpc.UnaryServerInterceptor](#GetMethodInterceptors)
* [func GetStreamInterceptors(svc interface{}, config CommonConfig) grpc.StreamServerInterceptor](#GetStreamInterceptors)
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
  * [func (m \*MiddlewareMapping) GetMiddlewaresFromURL(url string) []string](#MiddlewareMapping.GetMiddlewaresFromURL)
* [type Middlewareable](#Middlewareable)
* [type Optionable](#Optionable)
* [type StreamInterceptor](#StreamInterceptor)
* [type WhitelistedHeaders](#WhitelistedHeaders)

#### <a name="pkg-files">Package files</a>
[middleware.go](./middleware.go) [types.go](./types.go) [utils.go](./utils.go) 

## <a name="GetInterceptors">func</a> [GetInterceptors](./utils.go#L96)
``` go
func GetInterceptors(svc interface{}, config CommonConfig) grpc.UnaryServerInterceptor
```
GetInterceptors fetches interceptors from a given GRPC service

## <a name="GetInterceptorsWithMethodMiddlewares">func</a> [GetInterceptorsWithMethodMiddlewares](./utils.go#L106)
``` go
func GetInterceptorsWithMethodMiddlewares(svc interface{}, config CommonConfig, middlewares []string) grpc.UnaryServerInterceptor
```
GetInterceptorsWithMethodMiddlewares fetchs all middleware including those provided by method middlewares

## <a name="GetMethodInterceptors">func</a> [GetMethodInterceptors](./utils.go#L165)
``` go
func GetMethodInterceptors(svc interface{}, config CommonConfig, middlewares []string) []grpc.UnaryServerInterceptor
```
GetMethodInterceptors fetches all interceptors including method middlewares

## <a name="GetStreamInterceptors">func</a> [GetStreamInterceptors](./utils.go#L101)
``` go
func GetStreamInterceptors(svc interface{}, config CommonConfig) grpc.StreamServerInterceptor
```
GetStreamInterceptors fetches stream interceptors from a given GRPC service

## <a name="CommonConfig">type</a> [CommonConfig](./types.go#L79-L81)
``` go
type CommonConfig struct {
    NoDefaultInterceptors bool
}
```
CommonConfig is the config that is common across both http and grpc handlers

## <a name="Decodable">type</a> [Decodable](./types.go#L48-L51)
``` go
type Decodable interface {
    AddDecoder(serviceName, method string, decoder Decoder)
    AddDefaultDecoder(serviceName string, decoder Decoder)
}
```
Decodable interface that is implemented by a handler that supports custom HTTP decoder

## <a name="Decoder">type</a> [Decoder](./types.go#L39)
``` go
type Decoder func(ctx context.Context, w http.ResponseWriter, encodeError, endpointError error, respObject interface{})
```
Decoder is the function type needed for response decoders

## <a name="Encodeable">type</a> [Encodeable](./types.go#L42-L45)
``` go
type Encodeable interface {
    AddEncoder(serviceName, method string, httpMethod []string, path string, encoder Encoder)
    AddDefaultEncoder(serviceName string, encoder Encoder)
}
```
Encodeable interface that is implemented by a handler that supports custom HTTP encoder

## <a name="Encoder">type</a> [Encoder](./types.go#L36)
``` go
type Encoder func(req *http.Request, reqObject interface{}) error
```
Encoder is the function type needed for request encoders

## <a name="GRPCMethodHandler">type</a> [GRPCMethodHandler](./types.go#L13)
``` go
type GRPCMethodHandler func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error)
```
GRPCMethodHandler is the method type as defined in grpc-go

## <a name="HTTPHandler">type</a> [HTTPHandler](./types.go#L64)
``` go
type HTTPHandler func(http.ResponseWriter, *http.Request) bool
```
HTTPHandler is the function that handles HTTP request

## <a name="HTTPInterceptor">type</a> [HTTPInterceptor](./types.go#L59-L61)
``` go
type HTTPInterceptor interface {
    AddHTTPHandler(serviceName, method string, path string, handler HTTPHandler)
}
```
HTTPInterceptor allows intercepting an HTTP connection

## <a name="Handler">type</a> [Handler](./types.go#L67-L71)
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
    // gets an array of Unary Server Interceptors
    GetInterceptors() []grpc.UnaryServerInterceptor
}
```
Interceptor interface when implemented by a service allows that service to provide custom interceptors

## <a name="MiddlewareMapping">type</a> [MiddlewareMapping](./middleware.go#L14-L16)
``` go
type MiddlewareMapping struct {
    // contains filtered or unexported fields
}
```
MiddlewareMapping stores mapping between service,method and middlewares

### <a name="NewMiddlewareMapping">func</a> [NewMiddlewareMapping](./middleware.go#L9)
``` go
func NewMiddlewareMapping() *MiddlewareMapping
```
NewMiddlewareMapping returns a new MiddlewareMapping

### <a name="MiddlewareMapping.AddMiddleware">func</a> (\*MiddlewareMapping) [AddMiddleware](./middleware.go#L50)
``` go
func (m *MiddlewareMapping) AddMiddleware(service, method string, middlewares ...string)
```
AddMiddleware adds middleware to a service, method

### <a name="MiddlewareMapping.GetMiddlewares">func</a> (\*MiddlewareMapping) [GetMiddlewares](./middleware.go#L37)
``` go
func (m *MiddlewareMapping) GetMiddlewares(service, method string) []string
```
GetMiddlewares fetches all middlewares for a specific service,method

### <a name="MiddlewareMapping.GetMiddlewaresFromURL">func</a> (\*MiddlewareMapping) [GetMiddlewaresFromURL](./middleware.go#L31)
``` go
func (m *MiddlewareMapping) GetMiddlewaresFromURL(url string) []string
```
GetMiddlewaresFromURL fetches all middleware for a specific URL

## <a name="Middlewareable">type</a> [Middlewareable](./types.go#L74-L76)
``` go
type Middlewareable interface {
    AddMiddleware(serviceName, method string, middleware ...string)
}
```
Middlewareable implemets support for method specific middleware

## <a name="Optionable">type</a> [Optionable](./types.go#L54-L56)
``` go
type Optionable interface {
    AddOption(ServiceName, method, option string)
}
```
Optionable interface that is implemented by a handler that support custom Orion options

## <a name="StreamInterceptor">type</a> [StreamInterceptor](./types.go#L22-L25)
``` go
type StreamInterceptor interface {
    // gets an array of Stream Server Interceptors
    GetStreamInterceptors() []grpc.StreamServerInterceptor
}
```
StreamInterceptor interface when implemented by a service allows that service to provide custom stream interceptors

## <a name="WhitelistedHeaders">type</a> [WhitelistedHeaders](./types.go#L28-L33)
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