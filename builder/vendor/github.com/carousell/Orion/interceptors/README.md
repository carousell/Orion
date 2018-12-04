# interceptors
`import "github.com/carousell/Orion/interceptors"`

* [Overview](#pkg-overview)
* [Imported Packages](#pkg-imports)
* [Index](#pkg-index)

## <a name="pkg-overview">Overview</a>

## <a name="pkg-imports">Imported Packages</a>

- [github.com/afex/hystrix-go/hystrix](https://godoc.org/github.com/afex/hystrix-go/hystrix)
- [github.com/carousell/Orion/orion/modifiers](./../orion/modifiers)
- [github.com/carousell/Orion/utils](./../utils)
- [github.com/carousell/Orion/utils/errors](./../utils/errors)
- [github.com/carousell/Orion/utils/errors/notifier](./../utils/errors/notifier)
- [github.com/carousell/Orion/utils/log](./../utils/log)
- [github.com/carousell/Orion/utils/log/loggers](./../utils/log/loggers)
- [github.com/grpc-ecosystem/go-grpc-middleware](https://godoc.org/github.com/grpc-ecosystem/go-grpc-middleware)
- [github.com/grpc-ecosystem/go-grpc-middleware/tags](https://godoc.org/github.com/grpc-ecosystem/go-grpc-middleware/tags)
- [github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing](https://godoc.org/github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing)
- [github.com/grpc-ecosystem/go-grpc-prometheus](https://godoc.org/github.com/grpc-ecosystem/go-grpc-prometheus)
- [github.com/newrelic/go-agent](https://godoc.org/github.com/newrelic/go-agent)
- [google.golang.org/grpc](https://godoc.org/google.golang.org/grpc)

## <a name="pkg-index">Index</a>
* [Variables](#pkg-variables)
* [func DebugLoggingInterceptor() grpc.UnaryServerInterceptor](#DebugLoggingInterceptor)
* [func DefaultClientInterceptor(address string) grpc.UnaryClientInterceptor](#DefaultClientInterceptor)
* [func DefaultClientInterceptors(address string) []grpc.UnaryClientInterceptor](#DefaultClientInterceptors)
* [func DefaultInterceptors() []grpc.UnaryServerInterceptor](#DefaultInterceptors)
* [func DefaultStreamInterceptors() []grpc.StreamServerInterceptor](#DefaultStreamInterceptors)
* [func GRPCClientInterceptor() grpc.UnaryClientInterceptor](#GRPCClientInterceptor)
* [func HystrixClientInterceptor() grpc.UnaryClientInterceptor](#HystrixClientInterceptor)
* [func NewRelicClientInterceptor(address string) grpc.UnaryClientInterceptor](#NewRelicClientInterceptor)
* [func NewRelicInterceptor() grpc.UnaryServerInterceptor](#NewRelicInterceptor)
* [func ResponseTimeLoggingInterceptor() grpc.UnaryServerInterceptor](#ResponseTimeLoggingInterceptor)
* [func ServerErrorInterceptor() grpc.UnaryServerInterceptor](#ServerErrorInterceptor)
* [func WithHystrixName(name string) clientOption](#WithHystrixName)

#### <a name="pkg-files">Package files</a>
[documentations.go](./documentations.go) [interceptors.go](./interceptors.go) [options.go](./options.go) 

## <a name="pkg-variables">Variables</a>
``` go
var (
    //FilterMethods is the list of methods that are filtered by default
    FilterMethods = []string{"Healthcheck"}
)
```

## <a name="DebugLoggingInterceptor">func</a> [DebugLoggingInterceptor](./interceptors.go#L74)
``` go
func DebugLoggingInterceptor() grpc.UnaryServerInterceptor
```
DebugLoggingInterceptor is the interceptor that logs all request/response from a handler

## <a name="DefaultClientInterceptor">func</a> [DefaultClientInterceptor](./interceptors.go#L69)
``` go
func DefaultClientInterceptor(address string) grpc.UnaryClientInterceptor
```
DefaultClientInterceptor are the set of default interceptors that should be applied to all client calls

## <a name="DefaultClientInterceptors">func</a> [DefaultClientInterceptors](./interceptors.go#L51)
``` go
func DefaultClientInterceptors(address string) []grpc.UnaryClientInterceptor
```
DefaultClientInterceptors are the set of default interceptors that should be applied to all client calls

## <a name="DefaultInterceptors">func</a> [DefaultInterceptors](./interceptors.go#L39)
``` go
func DefaultInterceptors() []grpc.UnaryServerInterceptor
```
DefaultInterceptors are the set of default interceptors that are applied to all Orion methods

## <a name="DefaultStreamInterceptors">func</a> [DefaultStreamInterceptors](./interceptors.go#L60)
``` go
func DefaultStreamInterceptors() []grpc.StreamServerInterceptor
```
DefaultStreamInterceptors are the set of default interceptors that should be applied to all Orion streams

## <a name="GRPCClientInterceptor">func</a> [GRPCClientInterceptor](./interceptors.go#L154)
``` go
func GRPCClientInterceptor() grpc.UnaryClientInterceptor
```
GRPCClientInterceptor is the interceptor that intercepts all cleint requests and adds tracing info to them

## <a name="HystrixClientInterceptor">func</a> [HystrixClientInterceptor](./interceptors.go#L159)
``` go
func HystrixClientInterceptor() grpc.UnaryClientInterceptor
```
HystrixClientInterceptor is the interceptor that intercepts all cleint requests and adds hystrix info to them

## <a name="NewRelicClientInterceptor">func</a> [NewRelicClientInterceptor](./interceptors.go#L141)
``` go
func NewRelicClientInterceptor(address string) grpc.UnaryClientInterceptor
```
NewRelicClientInterceptor intercepts all client actions and reports them to newrelic

## <a name="NewRelicInterceptor">func</a> [NewRelicInterceptor](./interceptors.go#L98)
``` go
func NewRelicInterceptor() grpc.UnaryServerInterceptor
```
NewRelicInterceptor intercepts all server actions and reports them to newrelic

## <a name="ResponseTimeLoggingInterceptor">func</a> [ResponseTimeLoggingInterceptor](./interceptors.go#L84)
``` go
func ResponseTimeLoggingInterceptor() grpc.UnaryServerInterceptor
```
ResponseTimeLoggingInterceptor logs response time for each request on server

## <a name="ServerErrorInterceptor">func</a> [ServerErrorInterceptor](./interceptors.go#L117)
``` go
func ServerErrorInterceptor() grpc.UnaryServerInterceptor
```
ServerErrorInterceptor intercepts all server actions and reports them to error notifier

## <a name="WithHystrixName">func</a> [WithHystrixName](./options.go#L24)
``` go
func WithHystrixName(name string) clientOption
```
WithHystrixName changes the hystrix name to be used in the client interceptors

- - -
Generated by [godoc2ghmd](https://github.com/GandalfUK/godoc2ghmd)