# orion
`import "github.com/carousell/Orion/orion"`

* [Overview](#pkg-overview)
* [Imported Packages](#pkg-imports)
* [Index](#pkg-index)

## <a name="pkg-overview">Overview</a>
Package orion is a small lightwieght framework written around grpc with the aim to shorten time to build microservices

Source code for Orion can be found at <a href="https://github.com/carousell/Orion">https://github.com/carousell/Orion</a>

It is derived from 'Framework' a small microservices framework writen and used inside <a href="https://carousell.com">https://carousell.com</a>, It comes with a number of sensible defaults such as zipkin tracing, hystrix, live reload of configuration, etc.

### Whats Incuded
It comes included with

### Getting Started
First follow the install guide at <a href="https://github.com/carousell/Orion/blob/master/INSTALL.md">https://github.com/carousell/Orion/blob/master/INSTALL.md</a>

TODO -
doc gen

## <a name="pkg-imports">Imported Packages</a>

- [github.com/afex/hystrix-go/hystrix](https://godoc.org/github.com/afex/hystrix-go/hystrix)
- [github.com/carousell/Orion/orion/handlers](./handlers)
- [github.com/carousell/Orion/utils](./../utils)
- [github.com/carousell/Orion/utils/httptripper](./../utils/httptripper)
- [github.com/carousell/Orion/utils/listenerutils](./../utils/listenerutils)
- [github.com/go-kit/kit/log](https://godoc.org/github.com/go-kit/kit/log)
- [github.com/grpc-ecosystem/go-grpc-prometheus](https://godoc.org/github.com/grpc-ecosystem/go-grpc-prometheus)
- [github.com/newrelic/go-agent](https://godoc.org/github.com/newrelic/go-agent)
- [github.com/opentracing/opentracing-go](https://godoc.org/github.com/opentracing/opentracing-go)
- [github.com/openzipkin/zipkin-go-opentracing](https://godoc.org/github.com/openzipkin/zipkin-go-opentracing)
- [github.com/prometheus/client_golang/prometheus/promhttp](https://godoc.org/github.com/prometheus/client_golang/prometheus/promhttp)
- [github.com/spf13/viper](https://godoc.org/github.com/spf13/viper)
- [google.golang.org/grpc](https://godoc.org/google.golang.org/grpc)

## <a name="pkg-index">Index</a>
* [Constants](#pkg-constants)
* [Variables](#pkg-variables)
* [func AddConfigPath(path ...string)](#AddConfigPath)
* [func RegisterEncoder(svr Server, serviceName, method, httpMethod, path string, encoder handlers.Encoder)](#RegisterEncoder)
* [func ResetConfigPath()](#ResetConfigPath)
* [type Config](#Config)
  * [func BuildDefaultConfig(name string) Config](#BuildDefaultConfig)
* [type DefaultServerImpl](#DefaultServerImpl)
  * [func (d \*DefaultServerImpl) AddEncoder(serviceName, method, httpMethod string, path string, encoder handlers.Encoder)](#DefaultServerImpl.AddEncoder)
  * [func (d \*DefaultServerImpl) AddInitializers(ins ...Initializer)](#DefaultServerImpl.AddInitializers)
  * [func (d \*DefaultServerImpl) Fetch(key string) (value interface{}, found bool)](#DefaultServerImpl.Fetch)
  * [func (d \*DefaultServerImpl) GetConfig() map[string]interface{}](#DefaultServerImpl.GetConfig)
  * [func (d \*DefaultServerImpl) GetOrionConfig() Config](#DefaultServerImpl.GetOrionConfig)
  * [func (d \*DefaultServerImpl) RegisterService(sd \*grpc.ServiceDesc, sf ServiceFactory) error](#DefaultServerImpl.RegisterService)
  * [func (d \*DefaultServerImpl) Start()](#DefaultServerImpl.Start)
  * [func (d \*DefaultServerImpl) Stop(timeout time.Duration) error](#DefaultServerImpl.Stop)
  * [func (d \*DefaultServerImpl) Store(key string, value interface{})](#DefaultServerImpl.Store)
  * [func (d \*DefaultServerImpl) Wait() error](#DefaultServerImpl.Wait)
* [type Encoder](#Encoder)
* [type HystrixConfig](#HystrixConfig)
  * [func BuildDefaultHystrixConfig() HystrixConfig](#BuildDefaultHystrixConfig)
* [type Initializer](#Initializer)
  * [func HTTTPZipkinInitializer() Initializer](#HTTTPZipkinInitializer)
  * [func HystrixInitializer() Initializer](#HystrixInitializer)
  * [func NewRelicInitializer() Initializer](#NewRelicInitializer)
  * [func PprofInitializer() Initializer](#PprofInitializer)
  * [func PrometheusInitializer() Initializer](#PrometheusInitializer)
  * [func ZipkinInitializer() Initializer](#ZipkinInitializer)
* [type NewRelicConfig](#NewRelicConfig)
  * [func BuildDefaultNewRelicConfig() NewRelicConfig](#BuildDefaultNewRelicConfig)
* [type PostInitializer](#PostInitializer)
* [type PreInitializer](#PreInitializer)
* [type Server](#Server)
  * [func GetDefaultServer(name string) Server](#GetDefaultServer)
  * [func GetDefaultServerWithConfig(config Config) Server](#GetDefaultServerWithConfig)
* [type ServiceFactory](#ServiceFactory)
* [type ZipkinConfig](#ZipkinConfig)
  * [func BuildDefaultZipkinConfig() ZipkinConfig](#BuildDefaultZipkinConfig)

#### <a name="pkg-files">Package files</a>
[config.go](./config.go) [core.go](./core.go) [doc.go](./doc.go) [documetation.go](./documetation.go) [initializer.go](./initializer.go) [types.go](./types.go) [utils.go](./utils.go) 

## <a name="pkg-constants">Constants</a>
``` go
const (
    //BANNER is the orion banner text
    BANNER = `
  ___  ____  ___ ___  _   _
 / _ \|  _ \|_ _/ _ \| \ | |
| | | | |_) || | | | |  \| |
| |_| |  _ < | | |_| | |\  |
 \___/|_| \_\___\___/|_| \_|
                            `
)
```
``` go
const (
    // NRApp is the key for New Relic app object
    NRApp = "INIT:NR_APP"
)
```

## <a name="pkg-variables">Variables</a>
``` go
var (
    //DefaultInitializers are the initializers applied by orion as default
    DefaultInitializers = []Initializer{
        HystrixInitializer(),
        ZipkinInitializer(),
        NewRelicInitializer(),
        PrometheusInitializer(),
        PprofInitializer(),
        HTTTPZipkinInitializer(),
    }
)
```

## <a name="AddConfigPath">func</a> [AddConfigPath](./config.go#L153)
``` go
func AddConfigPath(path ...string)
```
AddConfigPath adds a config path from where orion tries to read config values

## <a name="RegisterEncoder">func</a> [RegisterEncoder](./utils.go#L9)
``` go
func RegisterEncoder(svr Server, serviceName, method, httpMethod, path string, encoder handlers.Encoder)
```
RegisterEncoder allows for registering an HTTP request encoder to arbitrary urls
Note: this is normally called from protoc-gen-orion autogenerated files

## <a name="ResetConfigPath">func</a> [ResetConfigPath](./config.go#L161)
``` go
func ResetConfigPath()
```
ResetConfigPath resets the configuration paths

## <a name="Config">type</a> [Config](./config.go#L17-L45)
``` go
type Config struct {
    //OrionServerName is the name of this orion server that is tracked
    OrionServerName string
    // GRPCOnly tells orion not to build HTTP/1.1 server and only initializes gRPC server
    GRPCOnly bool
    //HTTPOnly tells orion not to build gRPC server and only initializes HTTP/1.1 server
    HTTPOnly bool
    // HTTPPort is the port to bind for HTTP requests
    HTTPPort string
    // GRPCPost id the port to bind for gRPC requests
    GRPCPort string
    //PprofPort is the port to use for pprof
    PProfport string
    // HotReload when set reloads the service when it recieves SIGHUP
    HotReload bool
    //EnableProtoURL adds gRPC generated urls in HTTP handler
    EnableProtoURL bool
    //EnablePrometheus enables prometheus metric for services on path '/metrics' on pprof port
    EnablePrometheus bool
    //EnablePrometheusHistograms enables request histograms for services
    //ref: https://github.com/grpc-ecosystem/go-grpc-prometheus#histograms
    EnablePrometheusHistogram bool
    //HystrixConfig is the configuration options for hystrix
    HystrixConfig HystrixConfig
    //ZipkinConfig is the configuration options for zipkin
    ZipkinConfig ZipkinConfig
    //NewRelicConfig is the configuration options for new relic
    NewRelicConfig NewRelicConfig
}
```
Config is the configuration used by Orion core

### <a name="BuildDefaultConfig">func</a> [BuildDefaultConfig](./config.go#L68)
``` go
func BuildDefaultConfig(name string) Config
```
BuildDefaultConfig builds a default config object for Orion

## <a name="DefaultServerImpl">type</a> [DefaultServerImpl](./core.go#L40-L52)
``` go
type DefaultServerImpl struct {
    // contains filtered or unexported fields
}
```
DefaultServerImpl provides a default implementation of orion.Server this can be embeded in custom orion.Server implementations

### <a name="DefaultServerImpl.AddEncoder">func</a> (\*DefaultServerImpl) [AddEncoder](./core.go#L73)
``` go
func (d *DefaultServerImpl) AddEncoder(serviceName, method, httpMethod string, path string, encoder handlers.Encoder)
```
AddEncoder is the implementation of handlers.Encodable

### <a name="DefaultServerImpl.AddInitializers">func</a> (\*DefaultServerImpl) [AddInitializers](./core.go#L301)
``` go
func (d *DefaultServerImpl) AddInitializers(ins ...Initializer)
```
AddInitializers adds the initializers to orion server

### <a name="DefaultServerImpl.Fetch">func</a> (\*DefaultServerImpl) [Fetch](./core.go#L67)
``` go
func (d *DefaultServerImpl) Fetch(key string) (value interface{}, found bool)
```
Fetch fetches values for use by initializers

### <a name="DefaultServerImpl.GetConfig">func</a> (\*DefaultServerImpl) [GetConfig](./core.go#L312)
``` go
func (d *DefaultServerImpl) GetConfig() map[string]interface{}
```
GetConfig returns current config as parsed from the file/defaults

### <a name="DefaultServerImpl.GetOrionConfig">func</a> (\*DefaultServerImpl) [GetOrionConfig](./core.go#L88)
``` go
func (d *DefaultServerImpl) GetOrionConfig() Config
```
GetOrionConfig returns current orion config
NOTE: this config can not be modifies

### <a name="DefaultServerImpl.RegisterService">func</a> (\*DefaultServerImpl) [RegisterService](./core.go#L266)
``` go
func (d *DefaultServerImpl) RegisterService(sd *grpc.ServiceDesc, sf ServiceFactory) error
```
RegisterService registers a service from a generated proto file
Note: this is only callled from code generated by orion plugin

### <a name="DefaultServerImpl.Start">func</a> (\*DefaultServerImpl) [Start](./core.go#L222)
``` go
func (d *DefaultServerImpl) Start()
```
Start starts the orion server

### <a name="DefaultServerImpl.Stop">func</a> (\*DefaultServerImpl) [Stop](./core.go#L317)
``` go
func (d *DefaultServerImpl) Stop(timeout time.Duration) error
```
Stop stops the server

### <a name="DefaultServerImpl.Store">func</a> (\*DefaultServerImpl) [Store](./core.go#L55)
``` go
func (d *DefaultServerImpl) Store(key string, value interface{})
```
Store stores values for use by initializers

### <a name="DefaultServerImpl.Wait">func</a> (\*DefaultServerImpl) [Wait](./core.go#L259)
``` go
func (d *DefaultServerImpl) Wait() error
```
Wait waits for all the serving servers to quit

## <a name="Encoder">type</a> [Encoder](./types.go#L69)
``` go
type Encoder = handlers.Encoder
```
Encoder is the function type needed for request encoders

## <a name="HystrixConfig">type</a> [HystrixConfig](./config.go#L48-L53)
``` go
type HystrixConfig struct {
    //Port is the port to start hystrix stream handler on
    Port string
    //CommandConfig is configuration for individual commands
    CommandConfig map[string]hystrix.CommandConfig
}
```
HystrixConfig is configuration used by hystrix

### <a name="BuildDefaultHystrixConfig">func</a> [BuildDefaultHystrixConfig](./config.go#L88)
``` go
func BuildDefaultHystrixConfig() HystrixConfig
```
BuildDefaultHystrixConfig builds a default config for hystrix

## <a name="Initializer">type</a> [Initializer](./types.go#L45-L48)
``` go
type Initializer interface {
    Init(svr Server) error
    ReInit(svr Server) error
}
```
Initializer is the interface needed to be implemented by custom initializers

### <a name="HTTTPZipkinInitializer">func</a> [HTTTPZipkinInitializer](./initializer.go#L66)
``` go
func HTTTPZipkinInitializer() Initializer
```
HTTPZipkinInitializer returns an Initializer implementation for httptripper which appends zipkin trace info to all outgoing HTTP requests

### <a name="HystrixInitializer">func</a> [HystrixInitializer](./initializer.go#L41)
``` go
func HystrixInitializer() Initializer
```
HystrixInitializer returns a Initializer implementation for Hystrix

### <a name="NewRelicInitializer">func</a> [NewRelicInitializer](./initializer.go#L51)
``` go
func NewRelicInitializer() Initializer
```
NewRelicInitializer returns a Initializer implementation for NewRelic

### <a name="PprofInitializer">func</a> [PprofInitializer](./initializer.go#L61)
``` go
func PprofInitializer() Initializer
```
PprofInitializer returns a Initializer implementation for Pprof

### <a name="PrometheusInitializer">func</a> [PrometheusInitializer](./initializer.go#L56)
``` go
func PrometheusInitializer() Initializer
```
PrometheusInitializer returns a Initializer implementation for Prometheus

### <a name="ZipkinInitializer">func</a> [ZipkinInitializer](./initializer.go#L46)
``` go
func ZipkinInitializer() Initializer
```
ZipkinInitializer returns a Initializer implementation for Zipkin

## <a name="NewRelicConfig">type</a> [NewRelicConfig](./config.go#L62-L65)
``` go
type NewRelicConfig struct {
    APIKey      string
    ServiceName string
}
```
NewRelicConfig is the configuration for newrelic

### <a name="BuildDefaultNewRelicConfig">func</a> [BuildDefaultNewRelicConfig](./config.go#L103)
``` go
func BuildDefaultNewRelicConfig() NewRelicConfig
```
BuildDefaultNewRelicConfig builds a default config for newrelic

## <a name="PostInitializer">type</a> [PostInitializer](./types.go#L64-L66)
``` go
type PostInitializer interface {
    PostInit()
}
```
PostInitializer is the interface that needs to implemented by client for any custom code that runs after all other initializer

## <a name="PreInitializer">type</a> [PreInitializer](./types.go#L59-L61)
``` go
type PreInitializer interface {
    PreInit()
}
```
PreInitializer is the interface that needs to implemented by client for any custom code that runs before all other initializer

## <a name="Server">type</a> [Server](./types.go#L23-L42)
``` go
type Server interface {
    //Start starts the orion server, this is non blocking call
    Start()
    //RegisterService registers the service to origin server
    RegisterService(sd *grpc.ServiceDesc, sf ServiceFactory) error
    //Wait waits for the Server loop to exit
    Wait() error
    //Stop stops the Server
    Stop(timeout time.Duration) error
    //GetOrionConfig returns current orion config
    GetOrionConfig() Config
    //GetConfig returns current config as parsed from the file/defaults
    GetConfig() map[string]interface{}
    //AddInitializers adds the initializers to orion server
    AddInitializers(ins ...Initializer)
    //Store stores values for use by initializers
    Store(key string, value interface{})
    //Fetch fetches values for use by initializers
    Fetch(key string) (value interface{}, found bool)
}
```
Server is the interface that needs to be implemented by any orion server
'DefaultServerImpl' should be enough for most users.

### <a name="GetDefaultServer">func</a> [GetDefaultServer](./core.go#L331)
``` go
func GetDefaultServer(name string) Server
```
GetDefaultServer returns a default server object that can be directly used to start orion server

### <a name="GetDefaultServerWithConfig">func</a> [GetDefaultServerWithConfig](./core.go#L336)
``` go
func GetDefaultServerWithConfig(config Config) Server
```
GetDefaultServerWithConfig returns a default server object that uses provided configuration

## <a name="ServiceFactory">type</a> [ServiceFactory](./types.go#L51-L56)
``` go
type ServiceFactory interface {
    // NewService function recieves the server obejct for which service has to be initialized
    NewService(Server) interface{}
    //DisposeService function disposes the service object
    DisposeService(svc interface{})
}
```
ServiceFactory is the interface that need to be implemented by client that provides with a new service object

## <a name="ZipkinConfig">type</a> [ZipkinConfig](./config.go#L56-L59)
``` go
type ZipkinConfig struct {
    //Addr is the address of the zipkin collector
    Addr string
}
```
ZipkinConfig is the configuration for the zipkin collector

### <a name="BuildDefaultZipkinConfig">func</a> [BuildDefaultZipkinConfig](./config.go#L96)
``` go
func BuildDefaultZipkinConfig() ZipkinConfig
```
BuildDefaultZipkinConfig builds a default config for zipkin

- - -
Generated by [godoc2ghmd](https://github.com/GandalfUK/godoc2ghmd)