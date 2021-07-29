package forwardheaders

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func forwardHeadersThroughCtx(ctx context.Context) context.Context {
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

	return ctx
}

// UnaryClientInterceptor forwards metadata from upstream to downstream
func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx = forwardHeadersThroughCtx(ctx)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// StreamClientInterceptor forwards metadata from upstream to downstream
func StreamClientInterceptor() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		ctx = forwardHeadersThroughCtx(ctx)
		return streamer(ctx, desc, cc, method, opts...)
	}
}

// UnaryServerInterceptor sets the allowlist for every request to protect its outgoing requests
func UnaryServerInterceptor(allowlist []string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		ctx = SetAllowList(ctx, allowlist)
		return handler(ctx, req)
	}
}

type streamServer struct {
	grpc.ServerStream
	ctx context.Context
}

func (ss *streamServer) Context() context.Context {
	if ss.ctx == nil {
		return ss.ServerStream.Context()
	}
	return ss.ctx
}

// StreamServerInterceptor sets the allowlist for every request to protect its outgoing requests
func StreamServerInterceptor(allowlist []string) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()
		ctx = SetAllowList(ctx, allowlist)
		newServer := &streamServer{
			ServerStream: ss,
			ctx:          ctx,
		}
		return handler(srv, newServer)
	}
}
