package loggers

import (
	"context"
)

type logsContext string

var (
	contextKey logsContext = "LogsContextKey"
)

//LogFields contains all fields that have to be added to logs
type LogFields map[string]interface{}

// Add or modify log fields
func (o LogFields) Add(key string, value interface{}) {
	if len(key) > 0 {
		o[key] = value
	}
}

// Del deletes a log field entry
func (o LogFields) Del(key string) {
	delete(o, key)
}

//AddToLogContext adds log fields to context.
// Any info added here will be added to all logs using this context
func AddToLogContext(ctx context.Context, key string, value interface{}) context.Context {
	data := FromContext(ctx)
	if data == nil {
		ctx = context.WithValue(ctx, contextKey, make(LogFields))
		data = FromContext(ctx)
	}
	m := ctx.Value(contextKey)
	if data, ok := m.(LogFields); ok {
		data.Add(key, value)
	}
	return ctx
}

//FromContext fetchs log fields from provided context
func FromContext(ctx context.Context) LogFields {
	if ctx == nil {
		return nil
	}
	if h := ctx.Value(contextKey); h != nil {
		if logData, ok := h.(LogFields); ok {
			return logData
		}
	}
	return nil
}
