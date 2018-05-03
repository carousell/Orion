package handlers

import (
	"context"
	"log"
	"strings"

	"github.com/carousell/Orion/interceptors"
	"github.com/carousell/Orion/orion/modifiers"
	"github.com/carousell/Orion/utils/headers"
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

func getInterceptors(svc interface{}, config CommonConfig) grpc.UnaryServerInterceptor {

	opts := []grpc.UnaryServerInterceptor{optionsInterceptor}

	if config.NoDefaultInterceptors {
		// only add logging interceptor
		opts = append(opts, interceptors.ResponseTimeLoggingInterceptor())
	} else {
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

// grpcInterceptor acts as default interceptor for gprc and applies service specific interceptors based on implementation
func grpcInterceptor(config CommonConfig) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// fetch interceptors from the service implementation and apply
		interceptor := getInterceptors(info.Server, config)
		return interceptor(ctx, req, info, handler)
	}
}

func processWhitelist(data map[string][]string, allowedKeys []string) map[string][]string {
	whitelistedMap := make(map[string][]string)
	whitelistedKeys := make(map[string]bool)

	for _, k := range allowedKeys {
		whitelistedKeys[strings.ToLower(k)] = true
	}

	for k, v := range data {
		if _, found := whitelistedKeys[strings.ToLower(k)]; found {
			whitelistedMap[k] = v
		} else {
			log.Println("warning", "rejected headers not in whitelist", k, v)
		}
	}

	return whitelistedMap
}

//ContentTypeFromHeaders searches for a matching content type
func ContentTypeFromHeaders(ctx context.Context) string {
	hdrs := headers.RequestHeadersFromContext(ctx)
	if values, found := hdrs["Accept"]; found {
		for _, v := range values {
			if t, ok := ContentTypeMap[v]; ok {
				return t
			}
		}
	}
	if values, found := hdrs["Content-Type"]; found {
		for _, v := range values {
			if t, ok := ContentTypeMap[v]; ok {
				return t
			}
		}
	}
	return ""
}
