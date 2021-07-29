package interceptors

import (
	"context"
	"strings"

	"github.com/carousell/Orion/utils/forwardheaders"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
)

var (
	//FilterMethods is the list of methods that are filtered by default
	FilterMethods = []string{"Healthcheck", "HealthCheck"}
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
		PanicRecoveryInterceptor(),
	}
}

//DefaultClientInterceptors are the set of default interceptors that should be applied to all client calls
func DefaultClientInterceptors(address string) []grpc.UnaryClientInterceptor {
	return []grpc.UnaryClientInterceptor{
		grpc_retry.UnaryClientInterceptor(),
		GRPCClientInterceptor(),
		NewRelicClientInterceptor(address),
		HystrixClientInterceptor(),
		forwardheaders.UnaryClientInterceptor(),
	}
}

//DefaultStreamClientInterceptors are the set of default interceptors that should be applied to all client streaming calls
func DefaultStreamClientInterceptors() []grpc.StreamClientInterceptor {
	/*
		compare to DefaultClientInterceptors, we don't have hystrix and newrelic interceptors here
		because a stream call includes three parts: create connection, streaming, and close
		as an interceptor, it's not easy to differentiate them
	*/
	return []grpc.StreamClientInterceptor{
		grpc_retry.StreamClientInterceptor(),
		grpc_opentracing.StreamClientInterceptor(),
		forwardheaders.StreamClientInterceptor(),
	}
}

//DefaultStreamInterceptors are the set of default interceptors that should be applied to all Orion streams
func DefaultStreamInterceptors() []grpc.StreamServerInterceptor {
	return []grpc.StreamServerInterceptor{
		ResponseTimeLoggingStreamInterceptor(),
		grpc_ctxtags.StreamServerInterceptor(),
		grpc_opentracing.StreamServerInterceptor(),
		grpc_prometheus.StreamServerInterceptor,
		ServerErrorStreamInterceptor(),
	}
}

//DefaultClientInterceptor are the set of default interceptors that should be applied to all client calls
func DefaultClientInterceptor(address string) grpc.UnaryClientInterceptor {
	return grpc_middleware.ChainUnaryClient(DefaultClientInterceptors(address)...)
}

//DefaultStreamClientInterceptor are the set of default interceptors that should be applied to all client calls
func DefaultStreamClientInterceptor() grpc.StreamClientInterceptor {
	return grpc_middleware.ChainStreamClient(DefaultStreamClientInterceptors()...)
}
