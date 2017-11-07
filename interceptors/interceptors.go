package interceptors

import (
	"context"
	"fmt"
	"time"

	"github.com/coreos/rkt/tests/testutils/logger"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
)

func DebugLoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		fmt.Println(info, "requst", req)
		resp, err := handler(ctx, req)
		fmt.Println(info, "response", resp, err)
		return resp, err
	}
}

func ResponseTimeLoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func(begin time.Time) {
			logger.Log("method", info.FullMethod, "error", err, "took", time.Since(begin))
		}(time.Now())
		resp, err = handler(ctx, req)
		return resp, err
	}
}

func DefaultInterceptors() []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{
		ResponseTimeLoggingInterceptor(),
		grpc_opentracing.UnaryServerInterceptor(),
		grpc_prometheus.UnaryServerInterceptor,
	}
}
