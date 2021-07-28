package interceptors

import (
	"context"
	"time"

	"google.golang.org/grpc/metadata"

	"github.com/carousell/Orion/orion/modifiers"
	"github.com/carousell/Orion/utils/errors/notifier"
	"github.com/carousell/Orion/utils/log"
	"github.com/carousell/Orion/utils/log/loggers"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"google.golang.org/grpc"
)

// ResponseTimeLoggingStreamInterceptor logs response time for stream RPCs.
func ResponseTimeLoggingStreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer func(begin time.Time) {
			log.Info(stream.Context(), "method", info.FullMethod, "error", err, "took", time.Since(begin))
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
			ctx = loggers.AddToLogContext(ctx, "trace", traceID)
		}
		err = handler(srv, stream)
		if !modifiers.HasDontLogError(ctx) {
			notifier.Notify(err, ctx)
		}
		return err

	}
}

// ForwardMetadataInterceptor forwards metadata from upstream to downstream
func ForwardMetadataStreamClientInterceptor() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			// means that we have some incoming context values needed to pass through following services
			// e.g. api-gateway -> service1 -> service2
			for key, values := range md {
				for _, value := range values {
					ctx = metadata.AppendToOutgoingContext(ctx, key, value)
				}
			}
		}
		return streamer(ctx, desc, cc, method, opts...)
	}
}
