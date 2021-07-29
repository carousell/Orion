package forwardheaders

import (
	"context"
)

type contextKey string

var (
	allowlistKey contextKey = "AllowedForwardHeaders"
)

type hdr struct {
	allowlist []string
}

// AllowlistFromContext returns all allowed header keys passed in through context
func AllowlistFromContext(ctx context.Context) []string {
	if h := ctx.Value(allowlistKey); h != nil {
		if headers, ok := h.(*hdr); ok {
			return headers.allowlist
		}
	}
	return nil
}

// SetAllowList sets a list of allowed header keys in our package variable in through context
func SetAllowList(ctx context.Context, keys []string) context.Context {
	var h *hdr
	if ctxH := ctx.Value(allowlistKey); ctxH != nil {
		if convertedH, ok := ctxH.(*hdr); ok {
			h = convertedH
		}
	} else {
		h = &hdr{make([]string, 0)}
		ctx = context.WithValue(ctx, allowlistKey, h)
	}
	if h != nil && len(keys) > 0 {
		h.allowlist = keys
	}
	return ctx
}
