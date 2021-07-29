package forwardheaders

import (
	"context"
	"testing"

	"google.golang.org/grpc"
)

func TestUnaryClientInterceptor(t *testing.T) {
	ctx := context.Background()
	UnaryClientInterceptor()(ctx, "", new(interface{}), new(interface{}), new(grpc.ClientConn), func(invokerCtx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		// service function is wrapped during the chain of interceptors
		// so the context we will use is the one inside here
		// in order to test it, we do closure here to get the handlerCtx for validations below
		ctx = invokerCtx
		return nil
	})

	allowlist := AllowlistFromContext(ctx)
	if allowlist != nil {
		t.Errorf("allowlist is not nil by default")
	}

	ctx = SetAllowList(ctx, make([]string, 0))
	UnaryClientInterceptor()(ctx, "", new(interface{}), new(interface{}), new(grpc.ClientConn), func(invokerCtx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		// service function is wrapped during the chain of interceptors
		// so the context we will use is the one inside here
		// in order to test it, we do closure here to get the handlerCtx for validations below
		ctx = invokerCtx
		return nil
	})

	allowlist = AllowlistFromContext(ctx)
	if allowlist == nil {
		t.Errorf("allowlist is nil after set")
	}
}

func TestAllowlistFromContext(t *testing.T) {
	ctx := context.Background()
	if len(AllowlistFromContext(ctx)) != 0 {
		t.Errorf("allowlist shouldn't have value without set")
	}

	ctx = SetAllowList(ctx, []string{"platform"})

	if len(AllowlistFromContext(ctx)) != 1 {
		t.Errorf("the length of allowlist is wrong")
	}
}
