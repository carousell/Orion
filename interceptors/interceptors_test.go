package interceptors

import (
	"context"
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc/status"

	"google.golang.org/grpc/codes"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/carousell/Orion/v2/utils/errors"
	"google.golang.org/grpc"
)

func TestHystrixClientInterceptorPanicRecovery(t *testing.T) {
	panicErr := errors.New("test panic")
	invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		panic(panicErr)
	}

	err := HystrixClientInterceptor()(
		context.Background(),
		"method",
		nil,
		nil,
		nil,
		invoker,
	)

	if !strings.Contains(err.Error(), "panic") {
		t.Errorf("hystrixInterceptor doesn't recover panic properly, got:%v", err)
	}
}

func TestHystrixClientInterceptorTimeoutError(t *testing.T) {

	// default hystrix timeout is 1000 ms
	invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		time.Sleep(1100 * time.Millisecond)
		return nil
	}

	err := HystrixClientInterceptor()(
		context.Background(),
		"method",
		nil,
		nil,
		nil,
		invoker,
	)

	if !strings.Contains(err.Error(), hystrix.ErrTimeout.Error()) {
		t.Errorf("hystrixInterceptor doesn't propagate the hystrix timeout error properly, got:%v", err)
	}
}

func TestHystrixClientInterceptorOptionsCanIgnore(t *testing.T) {
	err1 := errors.New("hello world")
	tests := []struct {
		testName           string
		IsHystrixError     bool
		err                error
		ignorableErrors    []error
		ignorableGRPCCodes []codes.Code
	}{
		{
			testName:        "NonHystrixErrorByIgnorableErrors",
			IsHystrixError:  false,
			err:             err1,
			ignorableErrors: []error{err1},
		},
		{
			testName:           "NonHystrixErrorByIgnorableGRPCCodes",
			IsHystrixError:     false,
			err:                errors.NewWithStatus("test", status.New(codes.Canceled, "test")),
			ignorableGRPCCodes: []codes.Code{codes.Canceled},
		},
		{
			testName:        "HystrixError",
			IsHystrixError:  true,
			err:             errors.New("new error"),
			ignorableErrors: []error{errors.New("new error")},
		},
		{
			testName:        "HystrixNoError",
			IsHystrixError:  true, // meaning we can return the error (in this case, the error is nil, so it's fine)
			err:             nil,
			ignorableErrors: []error{errors.New("new error")},
		},
	}

	for _, test := range tests {
		options := hystrixOptions{
			cmdName: test.testName,
		}
		for _, opt := range []grpc.CallOption{
			WithHystrixIgnorableErrors(test.ignorableErrors...),
			WithHystrixIgnorableGRPCCodes(test.ignorableGRPCCodes...)} {
			if opt != nil {
				if o, ok := opt.(hystrixOption); ok {
					o.process(&options)
				}
			}
		}

		invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			return test.err
		}

		hystrixErr := HystrixClientInterceptor()(
			context.Background(),
			"method",
			nil,
			nil,
			nil,
			invoker,
			WithHystrixIgnorableErrors(test.ignorableErrors...),
			WithHystrixIgnorableGRPCCodes(test.ignorableGRPCCodes...),
		)

		if options.canIgnore(test.err) == test.IsHystrixError {
			t.Errorf("%s got problems on suppressing errors\n", test.testName)
		}

		if hystrixErr != test.err {
			t.Errorf("hystrixInterceptor doesn't return the expected error, got: %v, expect: %v", hystrixErr, test.err)
		}
	}
}
