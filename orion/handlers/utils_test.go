package handlers

import (
	"context"
	"google.golang.org/grpc"
	"reflect"
	"testing"
)

type service struct {
	interceptors []grpc.UnaryServerInterceptor
}

func NewMockService(interceptors ...grpc.UnaryServerInterceptor) *service {
	s := new(service)
	s.interceptors = interceptors
	return s
}

func (s *service) GetInterceptors() []grpc.UnaryServerInterceptor {
	return s.interceptors
}

func TestGetInterceptors(t *testing.T) {
	// this test currently verifies the Interceptors is executed at the same sequence that you passed in
	// gRPC's normal chainUnaryInterceptor is executing interceptor in a reverse order.
	// but Orion's chainUnaryServer ensure the interceptor execution order is normal (ensure this)

	sequenceArr := make([]int, 0) // for verifying GetInterceptors execution sequence
	expectedArr := []int{1, 2, 3}
	interceptors := []grpc.UnaryServerInterceptor{
		func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
			sequenceArr = append(sequenceArr, 1)
			return handler(ctx, req)
		},
		func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
			sequenceArr = append(sequenceArr, 2)
			return handler(ctx, req)
		},
		func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
			sequenceArr = append(sequenceArr, 3)
			return handler(ctx, req)
		},
	}
	svc := NewMockService(interceptors...)

	serverInterceptor := GetInterceptors(svc, CommonConfig{NoDefaultInterceptors: true})
	serverInterceptor(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "test"}, func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, nil
	}) // to trigger the interceptors

	if !reflect.DeepEqual(sequenceArr, expectedArr) {
		t.Errorf("execution sequence is not normal, expected: %v, got: %v\n", expectedArr, sequenceArr)
	}
}
