package handlers

import (
	"context"
	"testing"

	"github.com/carousell/Orion/utils/errors"
	"google.golang.org/grpc"
)

type testSvc struct {
	interceptors []grpc.UnaryServerInterceptor
}

func newTestSvc(interceptors []grpc.UnaryServerInterceptor) *testSvc {
	svc := &testSvc{interceptors: interceptors}
	return svc
}

func (svc *testSvc) GetInterceptors() []grpc.UnaryServerInterceptor {

	return svc.interceptors
}

func TestGetInterceptorForPanicHandling(t *testing.T) {
	svc := newTestSvc([]grpc.UnaryServerInterceptor{
		func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
			panic(errors.New("test panic"))
		},
	})
	i := GetInterceptorsWithMethodMiddlewares(svc, CommonConfig{NoDefaultInterceptors: false}, nil)

	i(context.Background(), nil, new(grpc.UnaryServerInfo), func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, nil
	})
}
