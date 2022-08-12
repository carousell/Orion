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
type ProtectedLogFields struct {
	Content LogFields
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
		ctx = context.WithValue(ctx, contextKey, &ProtectedLogFields{Content: make(LogFields)})
		data = fromContext(ctx)
	}
	m := ctx.Value(contextKey)
	if data, ok := m.(*ProtectedLogFields); ok {
		data.mtx.Lock()
		defer data.mtx.Unlock()
		// d := data.Content
		// fmt.Printf("Address %p\n", d)
		data.Content.Add(key, value)
	}
	return ctx
}

//FromContext fetchs log fields from provided context
func FromContext(ctx context.Context) LogFields {
	if ctx == nil {
		return nil
	}
	if h := ctx.Value(contextKey); h != nil {
		if plf, ok := h.(*ProtectedLogFields); ok {
			plf.mtx.RLock()
			defer plf.mtx.RUnlock()
			content := make(LogFields)
			for k, v := range plf.Content {
				content[k] = v
			}
			return content
		}
	}
	return nil
}

func fromContext(ctx context.Context) *ProtectedLogFields {
	if ctx == nil {
		return nil
	}
	if h := ctx.Value(contextKey); h != nil {
		if plf, ok := h.(*ProtectedLogFields); ok {
			return plf
		}
	}
	return nil
}
