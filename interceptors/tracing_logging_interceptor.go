package interceptors

import (
	"context"
	"fmt"
	"github.com/carousell/Orion/utils/errors"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"log"
)

type logCtxCarrier map[string]string

const (
	// Refer to https://zipkin.io/pages/instrumenting.html
	zipkinHeaderTraceId      = "x-b3-traceid"
	zipkinHeaderSpanId       = "x-b3-spanid"
	zipkinHeaderParentSpanId = "x-b3-parentspanid"
	zipkinHeaderSampled      = "x-b3-sampled"

	loggingCtxTraceKey = "trace_id"
)

func (l logCtxCarrier) Set(key, val string) {
	l[key] = val
}

//TraceIDLogInterceptor is a grpc server Interceptor that adds tracing info into logging context.
// Note: if the tracer is not initialized, nothing will be added to the logging context.
func TraceIDLogInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		fmt.Println(">>> interceptor start for method: ", info.FullMethod)
		ctx = appendTracingInfoToLoggingContext(ctx)
		fmt.Println(">>> interceptor end for method: ", info.FullMethod)
		resp, err = handler(ctx, req)
		return resp, err
	}
}

func appendTracingInfoToLoggingContext(ctx context.Context) context.Context {
	tracer := opentracing.GlobalTracer()
	var parentSpanCtx opentracing.SpanContext
	if parent := opentracing.SpanFromContext(ctx); parent != nil {
		parentSpanCtx = parent.Context()
	}

	lc := logCtxCarrier{}
	fmt.Println(">>>", "injection start")
	if err := tracer.Inject(parentSpanCtx, opentracing.HTTPHeaders, lc); err != nil {
		log.Println(errors.Wrap(err, "grpc_opentracing: failed serializing trace information: %"))
		return ctx
	}
	fmt.Println(">>>", "injection end")

	// if traceId is not present. skip adding to logging context
	traceId, ok := lc[zipkinHeaderTraceId]
	if !ok {
		return ctx
	}

	// if sampled is not present. skip adding to logging context
	sampleDecision, ok := lc[zipkinHeaderSampled]
	if !ok {
		return ctx
	}
	fmt.Println(">>> intercepted: trace: ", traceId, ", sampled: ", sampleDecision)

	// if not sampled. skip adding to logging context
	if sampleDecision != "1" && sampleDecision != "true" {
		return ctx
	}

	return ctx
}
