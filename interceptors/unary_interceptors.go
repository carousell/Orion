package interceptors

import (
	"context"
	"fmt"
	"time"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/carousell/Orion/orion/modifiers"
	"github.com/carousell/Orion/utils"
	"github.com/carousell/Orion/utils/errors"
	"github.com/carousell/Orion/utils/errors/notifier"
	"github.com/carousell/Orion/utils/log"
	"github.com/carousell/Orion/utils/log/loggers"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	newrelic "github.com/newrelic/go-agent"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

//DebugLoggingInterceptor is the interceptor that logs all request/response from a handler
func DebugLoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		fmt.Println(info, "requst", req)
		resp, err := handler(ctx, req)
		fmt.Println(info, "response", resp, "err", err)
		return resp, err
	}
}

//ResponseTimeLoggingInterceptor logs response time for each request on server
func ResponseTimeLoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		// dont log for HTTP request, let HTTP Handler manage it
		if !modifiers.IsHTTPRequest(ctx) {
			defer func(begin time.Time) {
				log.Info(ctx, "method", info.FullMethod, "error", err, "took", time.Since(begin))
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
			traceID := notifier.GetTraceId(ctx)
			t.Set("trace", traceID)
			ctx = loggers.AddToLogContext(ctx, "trace", traceID)
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

func PanicRecoveryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func(ctx context.Context) {
			// panic handler
			if r := recover(); r != nil {
				log.Error(ctx, "panic", r, "method", info.FullMethod)
				if e, ok := r.(error); ok {
					err = e
				} else {
					err = errors.New(fmt.Sprintf("panic: %s", r))
				}
				utils.FinishNRTransaction(ctx, err)
				notifier.NotifyWithLevel(err, "critical", info.FullMethod, ctx)
			}
		}(ctx)

		resp, err = handler(ctx, req)
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
		options := hystrixOptions{
			cmdName: method,
		}
		for _, opt := range opts {
			if opt != nil {
				if o, ok := opt.(hystrixOption); ok {
					o.process(&options)
				}
			}
		}
		newCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		var err error
		err = hystrix.Do(options.cmdName, func() (e error) {
			defer func() {
				if r := recover(); r != nil {
					err = errors.Wrap(fmt.Errorf("panic inside hystrix Method: %s, req: %v, reply: %v", method, req, reply), "Hystrix")
					log.Error(ctx, "panic", r, "method", method, "req", req, "reply", reply)
				}
			}()
			// error assigns back to the err object out of hystrix anyway
			defer notifier.NotifyOnPanic(newCtx, method)
			err = invoker(newCtx, method, req, reply, cc, opts...)
			if options.canIgnore(err) {
				return nil
			} else {
				return err
			}
		}, options.fallbackFunc)

		return err
	}
}

// ForwardMetadataInterceptor forwards metadata from upstream to downstream
func ForwardMetadataInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
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
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
