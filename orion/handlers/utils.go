package handlers

import (
	"context"

	"github.com/carousell/Orion/interceptors"
	"github.com/carousell/Orion/orion/modifiers"
	"github.com/carousell/Orion/utils/options"
	"google.golang.org/grpc"
)

// ChainUnaryServer creates a single interceptor out of a chain of many interceptors.
//
// Execution is done in left-to-right order, including passing of context.
// For example ChainUnaryServer(one, two, three) will execute one before two before three, and three
// will see context changes of one and two.
func chainUnaryServer(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	n := len(interceptors)

	if n > 1 {
		lastI := n - 1
		return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			var (
				chainHandler grpc.UnaryHandler
				curI         int
			)

			chainHandler = func(currentCtx context.Context, currentReq interface{}) (interface{}, error) {
				if curI == lastI {
					return handler(currentCtx, currentReq)
				}
				curI++
				return interceptors[curI](currentCtx, currentReq, info, chainHandler)
			}

			return interceptors[0](ctx, req, info, chainHandler)
		}
	}

	if n == 1 {
		return interceptors[0]
	}

	// n == 0; Dummy interceptor maintained for backward compatibility to avoid returning nil.
	return func(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		return handler(ctx, req)
	}
}

//GetInterceptors fetches interceptors from a given GRPC service
func GetInterceptors(svc interface{}, config CommonConfig) grpc.UnaryServerInterceptor {

	opts := []grpc.UnaryServerInterceptor{optionsInterceptor}

	if !config.NoDefaultInterceptors {
		// Add default interceptors
		opts = append(opts, interceptors.DefaultInterceptors()...)
	}

	interceptor, ok := svc.(Interceptor)
	if ok {
		opts = append(opts, interceptor.GetInterceptors()...)
	}

	return chainUnaryServer(opts...)
}

func optionsInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	ctx = options.AddToOptions(ctx, "", "")
	if !modifiers.IsHTTPRequest(ctx) {
		options.AddToOptions(ctx, modifiers.RequestGRPC, true)
	}
	return handler(ctx, req)
}
