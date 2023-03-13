package interceptors

import (
	"time"

	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"google.golang.org/grpc"

	"github.com/carousell/logging"
	"github.com/carousell/notifier"

	"github.com/carousell/Orion/v2/orion/modifiers"
)

// ResponseTimeLoggingStreamInterceptor logs response time for stream RPCs.
func ResponseTimeLoggingStreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer func(begin time.Time) {
			logging.Info(stream.Context(), "method", info.FullMethod, "error", err, "took", time.Since(begin))
		}(time.Now())
		err = handler(srv, stream)
		return err
	}
}

// ServerErrorStreamInterceptor intercepts server errors for stream RPCs and
// reports them to the error notifier.
func ServerErrorStreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		ctx := stream.Context()
		ctx = notifier.SetTraceId(ctx)
		t := grpc_ctxtags.Extract(ctx)
		if t != nil {
			traceID := notifier.GetTraceId(ctx)
			t.Set("trace", traceID)
			ctx = logging.AddToLogContext(ctx, "trace", traceID)
		}
		err = handler(srv, stream)
		if !modifiers.HasDontLogError(ctx) {
			notifier.Notify(err, ctx)
		}
		return err

	}
}
