package log

import (
	"context"
	"github.com/carousell/Orion/utils/log/loggers"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc/grpclog"
)

type loggingContext map[string][]string

func (l loggingContext) Set(key, val string) {
	l[key] = []string{val}
}

// ForeachKey is a opentracing.TextMapReader interface that extracts values.
func (l loggingContext) ForeachKey(callback func(key, val string) error) error {
	for k, vv := range l {
		for _, v := range vv {
			if err := callback(k, v); err != nil {
				return err
			}
		}
	}
	return nil
}

func AppendTracingInfoToLoggingContext(ctx context.Context) {
	tracer := opentracing.GlobalTracer()
	var parentSpanCtx opentracing.SpanContext
	if parent := opentracing.SpanFromContext(ctx); parent != nil {
		parentSpanCtx = parent.Context()
	}

	lc := loggingContext{}
	if err := tracer.Inject(parentSpanCtx, opentracing.HTTPHeaders, lc); err != nil {
		grpclog.Printf("grpc_opentracing: failed serializing trace information: %v", err)
	}

	for k, v := range lc {
		loggers.AddToLogContext(ctx, k, v)
	}

}