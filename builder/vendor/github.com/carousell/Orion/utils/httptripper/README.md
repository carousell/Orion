# httptripper
`import "github.com/carousell/Orion/utils/httptripper"`

* [Overview](#pkg-overview)
* [Imported Packages](#pkg-imports)
* [Index](#pkg-index)

## <a name="pkg-overview">Overview</a>
Package httptripper provides an implementation of http.RoundTripper that provides retries, popluates opentracing span info and hystrix circuit breaker.

### Setup
for most cases using the http.Client provided by the package is sufficient

	client := httptripper.NewHTTPClient(time.Millisecond * 500)

Note: If you are using a custom http.Client, then just wrap your custom http.Client using httptripper.WrapTripper

	tripper := httptripper.WrapTripper(client.Transport)
	client.Transport = tripper

### How To Use
Make sure you use httptripper.NewRequest to build http.Request, since http.NewRequest does not take context as parameter

	httpReq, err := httptripper.NewRequest(ctx, "TracingName", "GET", url, nil)

## <a name="pkg-imports">Imported Packages</a>

- [github.com/afex/hystrix-go/hystrix](https://godoc.org/github.com/afex/hystrix-go/hystrix)
- [github.com/carousell/Orion/utils/httptripper/retry](./retry)
- [github.com/carousell/Orion/utils/spanutils](./../spanutils)

## <a name="pkg-index">Index</a>
* [func GetRequestRetrier(req \*http.Request) retry.Retriable](#GetRequestRetrier)
* [func GetRequestTraceName(req \*http.Request) string](#GetRequestTraceName)
* [func NewHTTPClient(timeout time.Duration, options ...Option) \*http.Client](#NewHTTPClient)
* [func NewRequest(ctx context.Context, traceName, method, url string, body io.Reader) (\*http.Request, error)](#NewRequest)
* [func NewRequestWithRetrier(ctx context.Context, traceName string, retrier retry.Retriable, method, url string, body io.Reader) (\*http.Request, error)](#NewRequestWithRetrier)
* [func NewTripper(options ...Option) http.RoundTripper](#NewTripper)
* [func SetRequestRetrier(req \*http.Request, retrier retry.Retriable) \*http.Request](#SetRequestRetrier)
* [func SetRequestTraceName(req \*http.Request, traceName string) \*http.Request](#SetRequestTraceName)
* [func WrapTripper(base http.RoundTripper) http.RoundTripper](#WrapTripper)
* [type Option](#Option)
  * [func WithBaseTripper(base http.RoundTripper) Option](#WithBaseTripper)
  * [func WithHystrix(enabled bool) Option](#WithHystrix)
  * [func WithRetrier(retrier retry.Retriable) Option](#WithRetrier)
* [type OptionsData](#OptionsData)

#### <a name="pkg-files">Package files</a>
[httptripper.go](./httptripper.go) [types.go](./types.go) 

## <a name="GetRequestRetrier">func</a> [GetRequestRetrier](./httptripper.go#L199)
``` go
func GetRequestRetrier(req *http.Request) retry.Retriable
```
GetRequestRetrier fetches retrier to be used with this request

## <a name="GetRequestTraceName">func</a> [GetRequestTraceName](./httptripper.go#L181)
``` go
func GetRequestTraceName(req *http.Request) string
```
GetRequestTraceName fetches a trace name from HTTP request

## <a name="NewHTTPClient">func</a> [NewHTTPClient](./httptripper.go#L144)
``` go
func NewHTTPClient(timeout time.Duration, options ...Option) *http.Client
```
NewHTTPClient creates a new http.Client with default retry options and timeout

## <a name="NewRequest">func</a> [NewRequest](./httptripper.go#L156)
``` go
func NewRequest(ctx context.Context, traceName, method, url string, body io.Reader) (*http.Request, error)
```
NewRequest extends http.NewRequest with context and trace name

## <a name="NewRequestWithRetrier">func</a> [NewRequestWithRetrier](./httptripper.go#L165)
``` go
func NewRequestWithRetrier(ctx context.Context, traceName string, retrier retry.Retriable, method, url string, body io.Reader) (*http.Request, error)
```
NewRequestWithRetrier extends http.NewRequest with context, trace name and retrier

## <a name="NewTripper">func</a> [NewTripper](./httptripper.go#L130)
``` go
func NewTripper(options ...Option) http.RoundTripper
```
NewTripper returns a default tripper wrapped around http.DefaultTransport

## <a name="SetRequestRetrier">func</a> [SetRequestRetrier](./httptripper.go#L192)
``` go
func SetRequestRetrier(req *http.Request, retrier retry.Retriable) *http.Request
```
SetRequestRetrier sets the retrier to be used with this request

## <a name="SetRequestTraceName">func</a> [SetRequestTraceName](./httptripper.go#L174)
``` go
func SetRequestTraceName(req *http.Request, traceName string) *http.Request
```
SetRequestTraceName stores a trace name in a HTTP request

## <a name="WrapTripper">func</a> [WrapTripper](./httptripper.go#L125)
``` go
func WrapTripper(base http.RoundTripper) http.RoundTripper
```
WrapTripper wraps the base tripper with zipkin info

## <a name="Option">type</a> [Option](./types.go#L24)
``` go
type Option func(*OptionsData)
```
Option defines an options for Tripper

### <a name="WithBaseTripper">func</a> [WithBaseTripper](./httptripper.go#L210)
``` go
func WithBaseTripper(base http.RoundTripper) Option
```
WithBaseTripper updates the tripper to use the provided http.RoundTripper

### <a name="WithHystrix">func</a> [WithHystrix](./httptripper.go#L224)
``` go
func WithHystrix(enabled bool) Option
```
WithHystrix enables/disables use of hystrix

### <a name="WithRetrier">func</a> [WithRetrier](./httptripper.go#L217)
``` go
func WithRetrier(retrier retry.Retriable) Option
```
WithRetrier updates the tripper to use the provided retry.Retriable

## <a name="OptionsData">type</a> [OptionsData](./types.go#L17-L21)
``` go
type OptionsData struct {
    BaseTripper    http.RoundTripper
    HystrixEnabled bool
    Retrier        retry.Retriable
}
```
OptionsData is the data polulated by the options

- - -
Generated by [godoc2ghmd](https://github.com/GandalfUK/godoc2ghmd)