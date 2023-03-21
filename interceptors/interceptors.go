package interceptors

import (
	"google.golang.org/grpc"
)

//DefaultInterceptors are the set of default interceptors that are applied to all Orion methods
func DefaultInterceptors() []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{
		PanicRecoveryInterceptor(),
	}
}

//DefaultStreamInterceptors are the set of default interceptors that should be applied to all Orion streams
func DefaultStreamInterceptors() []grpc.StreamServerInterceptor {
	return []grpc.StreamServerInterceptor{}
}
