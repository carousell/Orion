package loggers

import (
	"context"
	"sync"
)

type logsContext string

var (
	contextKey logsContext = "LogsContextKey"
)

//LogFields contains all fields that have to be added to logs
type LogFields map[string]interface{}
type protectedLogFields struct {
	content LogFields
	mtx     sync.RWMutex
}

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
	data := fromContext(ctx)
	//Initialize if key doesn't exist
	if data == nil {
		data = &protectedLogFields{content: make(LogFields)}
		ctx = context.WithValue(ctx, contextKey, data)
	}
	data.mtx.Lock()
	defer data.mtx.Unlock()
	data.content.Add(key, value)
	return ctx
}

//FromContext fetchs log fields from provided context
func FromContext(ctx context.Context) LogFields {
	if plf := fromContext(ctx); plf != nil {
		plf.mtx.RLock()
		defer plf.mtx.RUnlock()
		content := make(LogFields)
		for k, v := range plf.content {
			content[k] = v
		}
		return content
	}
	return nil
}

func fromContext(ctx context.Context) *protectedLogFields {
	if ctx == nil {
		return nil
	}
	if h := ctx.Value(contextKey); h != nil {
		if plf, ok := h.(*protectedLogFields); ok {
			return plf
		}
	}
	return nil
}
