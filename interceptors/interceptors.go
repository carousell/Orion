package interceptors

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/carousell/Orion/utils"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	newrelic "github.com/newrelic/go-agent"
	"google.golang.org/grpc"
)

//DefaultInterceptors are the set of default interceptors that are applied to all Orion methods
func DefaultInterceptors() []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{
		ResponseTimeLoggingInterceptor(),
		NewRelicInterceptor(),
		grpc_opentracing.UnaryServerInterceptor(),
		grpc_prometheus.UnaryServerInterceptor,
	}
}

func DefaultClientInterceptors(address string) []grpc.UnaryClientInterceptor {
	return []grpc.UnaryClientInterceptor{
		GRPCClientInterceptor(),
		NewRelicClientInterceptor(address),
		HystrixClientInterceptor(),
	}
}

//DebugLoggingInterceptor is the interceptor that logs all request/response from a handler
func DebugLoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		fmt.Println(info, "requst", req)
		resp, err := handler(ctx, req)
		fmt.Println(info, "response", resp, err)
		return resp, err
	}
}

//ResponseTimeLoggingInterceptor logs response time for each request on server
func ResponseTimeLoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func(begin time.Time) {
			log.Println("method", info.FullMethod, "error", err, "took", time.Since(begin))
		}(time.Now())
		resp, err = handler(ctx, req)
		return resp, err
	}
}

func NewRelicInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		ctx = utils.StartNRTransaction(info.FullMethod, ctx, nil, nil)
		resp, err = handler(ctx, req)
		utils.FinishNRTransaction(ctx, err)
		return resp, err
	}
}

func NewRelicClientInterceptor(address string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		txn := utils.GetNewRelicTransactionFromContext(ctx)
		seg := newrelic.ExternalSegment{
			StartTime: newrelic.StartSegmentNow(txn),
			URL:       "http://" + address + "/" + method,
		}
		defer seg.End()
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

//GRPCCLientInterceptor is the interceptor that intercepts all cleint requests and adds tracing info to them
func GRPCClientInterceptor() grpc.UnaryClientInterceptor {
	return grpc_opentracing.UnaryClientInterceptor()
}

//HystrixClientInterceptor is the interceptor that intercepts all cleint requests and adds hystrix info to them
func HystrixClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		return hystrix.Do(method, func() error {
			return invoker(ctx, method, req, reply, cc, opts...)
		}, nil)
	}
}
