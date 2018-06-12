package interceptors

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/carousell/Orion/orion/modifiers"
	"github.com/carousell/Orion/utils"
	"github.com/carousell/Orion/utils/errors/notifier"
	"github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	newrelic "github.com/newrelic/go-agent"
	"google.golang.org/grpc"
)

var (
	//FilterMethods is the list of methods that are filtered by default
	FilterMethods = []string{"Healthcheck"}
)

func filterFromZipkin(ctx context.Context, fullMethodName string) bool {
	for _, name := range FilterMethods {
		if strings.Contains(fullMethodName, name) {
			return false
		}
	}
	return true
}

//DefaultInterceptors are the set of default interceptors that are applied to all Orion methods
func DefaultInterceptors() []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{
		ResponseTimeLoggingInterceptor(),
		grpc_ctxtags.UnaryServerInterceptor(),
		grpc_opentracing.UnaryServerInterceptor(grpc_opentracing.WithFilterFunc(filterFromZipkin)),
		grpc_prometheus.UnaryServerInterceptor,
		ServerErrorInterceptor(),
		NewRelicInterceptor(),
	}
}

//DefaultClientInterceptors are the set of default interceptors that should be applied to all client calls
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
		// dont log for HTTP request, let HTTP Handler manage it
		if !modifiers.IsHTTPRequest(ctx) {
			defer func(begin time.Time) {
				log.Println("method", info.FullMethod, "error", err, "took", time.Since(begin))
			}(time.Now())
		}
		resp, err = handler(ctx, req)
		return resp, err
	}
}

//NewRelicInterceptor intercepts all server actions and reports them to newrelic
func NewRelicInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		// dont log NR for HTTP request, let HTTP Handler manage it
		if modifiers.IsHTTPRequest(ctx) {
			return handler(ctx, req)
		}
		ctx = utils.StartNRTransaction(info.FullMethod, ctx, nil, nil)
		resp, err = handler(ctx, req)
		if modifiers.HasDontLogError(ctx) {
			// dont report error to NR
			utils.FinishNRTransaction(ctx, nil)
		} else {
			utils.FinishNRTransaction(ctx, err)
		}
		return resp, err
	}
}

//ServerErrorInterceptor intercepts all server actions and reports them to error notifier
func ServerErrorInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		// set trace id if not set
		ctx = notifier.SetTraceId(ctx)

		t := grpc_ctxtags.Extract(ctx)
		if t != nil {
			t.Set("trace", notifier.GetTraceId(ctx))
		}
		// dont log Error for HTTP request, let HTTP Handler manage it
		if modifiers.IsHTTPRequest(ctx) {
			return handler(ctx, req)
		}
		resp, err = handler(ctx, req)
		if !modifiers.HasDontLogError(ctx) {
			notifier.Notify(err, ctx)
		}
		return resp, err
	}
}

//NewRelicClientInterceptor intercepts all client actions and reports them to newrelic
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

//GRPCClientInterceptor is the interceptor that intercepts all cleint requests and adds tracing info to them
func GRPCClientInterceptor() grpc.UnaryClientInterceptor {
	return grpc_opentracing.UnaryClientInterceptor()
}

//HystrixClientInterceptor is the interceptor that intercepts all cleint requests and adds hystrix info to them
func HystrixClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		newCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		return hystrix.Do(method, func() error {
			return invoker(newCtx, method, req, reply, cc, opts...)
		}, nil)
	}
}
