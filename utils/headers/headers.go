package headers

import (
	"context"
	"net/http"
)

type contextKey string

var (
	requestHeadersKey  contextKey = "OrionRequestHeaders"
	responseHeadersKey contextKey = "OrionResponseHeaders"
)

type hdr struct {
	http.Header
}

//RequestHeadersFromContext returns all request headers passed in through context
func RequestHeadersFromContext(ctx context.Context) http.Header {
	if h := ctx.Value(requestHeadersKey); h != nil {
		if headers, ok := h.(*hdr); ok {
			return headers.Header
		}
	}
	return nil
}

//ResponseHeadersFromContext returns all response headers passed in through context
func ResponseHeadersFromContext(ctx context.Context) http.Header {
	if h := ctx.Value(responseHeadersKey); h != nil {
		if headers, ok := h.(*hdr); ok {
			return headers.Header
		}
	}
	return nil
}

//AddToRequestHeaders adds a request header to headers passed in through context
func AddToRequestHeaders(ctx context.Context, key string, value string) context.Context {
	h := RequestHeadersFromContext(ctx)
	if h == nil {
		ctx = context.WithValue(ctx, requestHeadersKey, &hdr{make(map[string][]string)})
	}
	h = RequestHeadersFromContext(ctx)
	if h != nil && key != "" {
		h.Add(key, value)
	}
	return ctx
}

//AddToResponseHeaders adds a response header to headers that will returned through context
func AddToResponseHeaders(ctx context.Context, key string, value string) context.Context {
	h := ResponseHeadersFromContext(ctx)
	if h == nil {
		ctx = context.WithValue(ctx, responseHeadersKey, &hdr{make(map[string][]string)})
	}
	h = ResponseHeadersFromContext(ctx)
	if h != nil && key != "" {
		h.Add(key, value)
	}
	return ctx
}
