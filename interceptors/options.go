package interceptors

import "google.golang.org/grpc"

type clientOption interface {
	grpc.CallOption
	process(*clientOptions)
}

type clientOptions struct {
	hystrixName string
	hystrixErrorSuppressor func(err error)error
}

type optionCarrier struct {
	grpc.EmptyCallOption
	processor func(*clientOptions)
}

func (h *optionCarrier) process(co *clientOptions) {
	h.processor(co)
}

//WithHystrixName changes the hystrix name to be used in the client interceptors
func WithHystrixName(name string) clientOption {
	return &optionCarrier{
		processor: func(co *clientOptions) {
			if name != "" {
				co.hystrixName = name
			}
		},
	}
}

// WithHystrixErrorSuppressor applies a function that you can write logics to suppress errors to hystrix interceptor
func WithHystrixErrorSuppressor(e func(err error) error) clientOption {
	return &optionCarrier{
		processor: func(co *clientOptions) {
			if e != nil {
				co.hystrixErrorSuppressor = e
			}
		},
	}
}
