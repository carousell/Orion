package options

import (
	"context"
	"strings"
)

type contextKey string

var (
	optionsKey contextKey = "OrionOptions"
)

// Options are request options passed from Orion to server
type Options map[string]interface{}

//FromContext fetchs options from provided context
func FromContext(ctx context.Context) Options {
	if h := ctx.Value(optionsKey); h != nil {
		if options, ok := h.(Options); ok {
			return options
		}
	}
	return nil
}

//AddToOptions adds options to context
func AddToOptions(ctx context.Context, key string, value interface{}) context.Context {
	h := FromContext(ctx)
	if h == nil {
		ctx = context.WithValue(ctx, optionsKey, make(Options))
	}
	h = FromContext(ctx)
	if h != nil && key != "" {
		h.Add(key, value)
	}
	return ctx
}

// Add to Options
func (o Options) Add(key string, value interface{}) {
	o[strings.ToLower(key)] = value
}

// Del an options
func (o Options) Del(key string) {
	delete(o, strings.ToLower(key))
}

//Get an options
func (o Options) Get(key string) (interface{}, bool) {
	value, found := o[strings.ToLower(key)]
	return value, found
}
