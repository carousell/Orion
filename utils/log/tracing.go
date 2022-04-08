package log

import (
	"context"
	"fmt"
	"github.com/carousell/Orion/utils/log/loggers"
	"github.com/opentracing/opentracing-go"
)

type loggingContext map[string][]string

func (l loggingContext) Set(key, val string) {
	if values, ok := l[key]; ok {
		l[key] = append(values, val)
	} else {
		l[key] = []string{val}
	}
}

func AppendTracingInfoToLoggingContext(ctx context.Context) {
	tracer := opentracing.GlobalTracer()
	var parentSpanCtx opentracing.SpanContext
	if parent := opentracing.SpanFromContext(ctx); parent != nil {
		parentSpanCtx = parent.Context()
	}
	fmt.Println(">>> AppendTracingInfoToLoggingContext: parentSpanCtx", parentSpanCtx)

	lc := loggingContext{}
	if err := tracer.Inject(parentSpanCtx, opentracing.TextMap, lc); err != nil {
		fmt.Printf("grpc_opentracing: failed serializing trace information: %v\n", err)
	}

	for k, v := range lc {
		fmt.Println(">>> AppendTracingInfoToLoggingContext: k ", k, ", v: ", v)
		loggers.AddToLogContext(ctx, k, v)
	}

}
