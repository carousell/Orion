package utils

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	opentracing "github.com/opentracing/opentracing-go"
	otext "github.com/opentracing/opentracing-go/ext"
	"google.golang.org/grpc"
)

// TraceServer returns a Middleware that wraps the `next` Endpoint in an
// OpenTracing Span called `operationName`.
//
// If `ctx` already has a Span, it is re-used and the operation name is
// overwritten. If `ctx` does not yet have a Span, one is created here.
func TraceServer(tracer opentracing.Tracer, operationName string) grpc.UnaryServerInterceptor {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			serverSpan := opentracing.SpanFromContext(ctx)
			if serverSpan == nil {
				// All we can do is create a new root span.
				serverSpan = tracer.StartSpan(operationName)
			} else {
				serverSpan.SetOperationName(operationName)
			}
			defer serverSpan.Finish()
			otext.SpanKindRPCServer.Set(serverSpan)
			ctx = opentracing.ContextWithSpan(ctx, serverSpan)
			return next(ctx, request)
		}
	}
}
