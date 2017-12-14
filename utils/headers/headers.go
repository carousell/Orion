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

func RequestHeadersFromContext(ctx context.Context) http.Header {
	if h := ctx.Value(requestHeadersKey); h != nil {
		if headers, ok := h.(*hdr); ok {
			return headers.Header
		}
	}
	return nil
}

func ResponseHeadersFromContext(ctx context.Context) http.Header {
	if h := ctx.Value(responseHeadersKey); h != nil {
		if headers, ok := h.(*hdr); ok {
			return headers.Header
		}
	}
	return nil
}

func AddToRequestHeaders(ctx context.Context, key string, value string) context.Context {
	h := RequestHeadersFromContext(ctx)
	if h == nil {
		ctx = context.WithValue(ctx, requestHeadersKey, &hdr{})
	}
	h = RequestHeadersFromContext(ctx)
	if h != nil && key != "" {
		h.Add(key, value)
	}
	return ctx
}

func AddToResponseHeaders(ctx context.Context, key string, value string) context.Context {
	h := ResponseHeadersFromContext(ctx)
	if h == nil {
		ctx = context.WithValue(ctx, responseHeadersKey, &hdr{})
	}
	h = ResponseHeadersFromContext(ctx)
	if h != nil && key != "" {
		h.Add(key, value)
	}
	return ctx
}
