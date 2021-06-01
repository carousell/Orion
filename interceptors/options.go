package interceptors

import (
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type hystrixOption interface {
	grpc.CallOption
	process(*hystrixOptions)
}

type hystrixOptions struct {
	cmdName            string
	ignorableErrors    []error
	ignorableGRPCCodes []codes.Code
	fallbackFunc       func(error) error
}

type optionCarrier struct {
	grpc.EmptyCallOption
	processor func(*hystrixOptions)
}

func (h *optionCarrier) process(co *hystrixOptions) {
	h.processor(co)
}

//WithHystrixName changes the hystrix name to be used in the client interceptors
func WithHystrixName(name string) hystrixOption {
	return &optionCarrier{
		processor: func(co *hystrixOptions) {
			if name != "" {
				co.cmdName = name
			}
		},
	}
}

func WithHystrixIgnorableErrors(e ...error) hystrixOption {
	return &optionCarrier{
		processor: func(co *hystrixOptions) {
			co.ignorableErrors = e
		},
	}
}

func WithHystrixIgnorableGRPCCodes(c ...codes.Code) hystrixOption {
	return &optionCarrier{
		processor: func(co *hystrixOptions) {
			co.ignorableGRPCCodes = c
		},
	}
}
func WithHystrixFallbackFunc(f func(error) error) hystrixOption {
	return &optionCarrier{
		processor: func(co *hystrixOptions) {
			co.fallbackFunc = f
		},
	}
}

func (ho *hystrixOptions) canIgnore(err error) bool {
	if err == nil {
		return false
	}

	if ge, ok := status.FromError(err); ok {
		for _, c := range ho.ignorableGRPCCodes {
			if ge.Code() == c {
				return true
			}
		}
	}
	for _, e := range ho.ignorableErrors {
		if errors.Cause(err) == errors.Cause(e) {
			return true
		}
	}
	return false
}
