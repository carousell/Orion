package headers

import (
	"context"
	"strings"
)

type contextKey string

var (
	requestHeadersKey  contextKey = "OrionRequestHeaders"
	responseHeadersKey contextKey = "OrionResponseHeaders"
)

type hdr struct {
	data map[string]string
}

func (h *hdr) Add(key string, value string) {
	if h.data == nil {
		h.data = make(map[string]string)
	}
	if strings.TrimSpace(key) != "" {
		h.data[key] = value
	}
}

func (h *hdr) Del(key string) {
	if h.data == nil {
		h.data = make(map[string]string)
	}

	if strings.TrimSpace(key) != "" {
		delete(h.data, key)
	}
}

func (h *hdr) Get(key string) string {
	if h.data == nil {
		h.data = make(map[string]string)
	}
	return h.data[key]
}

func (h *hdr) GetAll() map[string]string {
	return h.data
}

func RequestHeadersFromContext(ctx context.Context) Headers {
	if h := ctx.Value(requestHeadersKey); h != nil {
		if headers, ok := h.(*hdr); ok {
			return headers
		}
	}
	return nil
}

func ResponseHeadersFromContext(ctx context.Context) Headers {
	if h := ctx.Value(responseHeadersKey); h != nil {
		if headers, ok := h.(*hdr); ok {
			return headers
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
