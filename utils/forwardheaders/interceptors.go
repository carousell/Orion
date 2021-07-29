package forwardheaders

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// UnaryClientInterceptor forwards metadata from upstream to downstream
func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			// means that we have some incoming context values needed to pass through following services
			// e.g. api-gateway -> service1 -> service2
			allowlist := AllowlistFromContext(ctx)
			if allowlist == nil {
				for key, values := range md {
					for _, value := range values {
						ctx = metadata.AppendToOutgoingContext(ctx, key, value)
					}
				}
			} else {
				for _, key := range allowlist {
					if values, ok := md[key]; ok {
						for _, value := range values {
							ctx = metadata.AppendToOutgoingContext(ctx, key, value)
						}
					}
				}
			}
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
