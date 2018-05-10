# headers
`import "github.com/carousell/Orion/utils/headers"`

* [Overview](#pkg-overview)
* [Imported Packages](#pkg-imports)
* [Index](#pkg-index)

## <a name="pkg-overview">Overview</a>

## <a name="pkg-imports">Imported Packages</a>

No packages beyond the Go standard library are imported.

## <a name="pkg-index">Index</a>
* [func AddToRequestHeaders(ctx context.Context, key string, value string) context.Context](#AddToRequestHeaders)
* [func AddToResponseHeaders(ctx context.Context, key string, value string) context.Context](#AddToResponseHeaders)
* [func RequestHeadersFromContext(ctx context.Context) http.Header](#RequestHeadersFromContext)
* [func ResponseHeadersFromContext(ctx context.Context) http.Header](#ResponseHeadersFromContext)

#### <a name="pkg-files">Package files</a>
[headers.go](./headers.go) 

## <a name="AddToRequestHeaders">func</a> [AddToRequestHeaders](./headers.go#L40)
``` go
func AddToRequestHeaders(ctx context.Context, key string, value string) context.Context
```
AddToRequestHeaders adds a request header to headers passed in through context

## <a name="AddToResponseHeaders">func</a> [AddToResponseHeaders](./headers.go#L53)
``` go
func AddToResponseHeaders(ctx context.Context, key string, value string) context.Context
```
AddToResponseHeaders adds a response header to headers that will returned through context

## <a name="RequestHeadersFromContext">func</a> [RequestHeadersFromContext](./headers.go#L20)
``` go
func RequestHeadersFromContext(ctx context.Context) http.Header
```
RequestHeadersFromContext returns all request headers passed in through context

## <a name="ResponseHeadersFromContext">func</a> [ResponseHeadersFromContext](./headers.go#L30)
``` go
func ResponseHeadersFromContext(ctx context.Context) http.Header
```
ResponseHeadersFromContext returns all response headers passed in through context

- - -
Generated by [godoc2ghmd](https://github.com/GandalfUK/godoc2ghmd)