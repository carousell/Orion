package interceptors

import (
	"context"
	"github.com/afex/hystrix-go/hystrix"
	"github.com/carousell/Orion/utils/errors"
	"google.golang.org/grpc"
	"testing"
)

func TestHystrixClientInterceptor(t *testing.T) {
	suppressedErr := errors.New("test error")
	var isHystrixError bool
	errorSuppressor := func(e error)error{
		if e == suppressedErr {
			isHystrixError = false
			return nil
		}
		isHystrixError = true
		return e
	}

	tests := []struct{
		testName string
		IsHystrixError bool
		err error
	}{
		{
			testName: "HystrixError",
			IsHystrixError: true,
			err: errors.New("hello world"),
		},
		{
			testName: "NonHystrixError",
			IsHystrixError: false,
			err: suppressedErr,
		},
	}

	// not support concurrent testing yet
	for _, test:= range tests {
		// Test WithHystrixErrorSuppressor
		hystrixCmd := test.testName
		hystrix.ConfigureCommand(hystrixCmd, hystrix.CommandConfig{
			ErrorPercentThreshold: 0,
		})
		invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			return test.err
		}
		err := HystrixClientInterceptor()(
			context.Background(),
			"method",
			nil,
			nil,
			nil,
			invoker,
			WithHystrixName(hystrixCmd),
			WithHystrixErrorSuppressor(errorSuppressor),
		)

		if isHystrixError != test.IsHystrixError {
			t.Errorf("%s got problems on error suppressor\n", hystrixCmd)
		}
		if err != test.err {
			// we expect the same error as we define but it could be suppressed on hystrix
			t.Errorf("%s got different errors", hystrixCmd)
		}
	}
}
