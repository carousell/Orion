package httptripper

import (
	"net/http"
	"time"
)

type defaultRetry struct {
	option *RetryOptions
}

func (d *defaultRetry) ShouldRetry(retryConut int, req *http.Request, resp *http.Response, err error) bool {
	if resp != nil {
		// dont retry for anything less than 5XX
		if resp.StatusCode < 500 {
			return false
		}
	}
	if retryConut < d.option.MaxRetry {
		if d.option.RetryAllMethods {
			return true
		}
		if allowed, ok := d.option.Methods[req.Method]; ok {
			return allowed
		}
	}
	return false
}

func (d *defaultRetry) WaitDuration(retryConut int, req *http.Request, resp *http.Response, err error) time.Duration {
	return d.option.Delay
}

func NewRetry(options ...RetryOption) Retriable {
	r := &defaultRetry{
		option: &RetryOptions{},
	}
	for _, opt := range options {
		opt(r.option)
	}
	return r
}

func DefaultRetry() Retriable {
	return NewRetry(
		WithMaxRetry(3),
		WithRetryDelay(time.Second),
		WithRetryMethods(http.MethodGet, http.MethodOptions, http.MethodHead),
		WithRetryAllMethods(false),
	)
}

func WithMaxRetry(max int) RetryOption {
	return func(ro *RetryOptions) {
		ro.MaxRetry = max
	}
}

func WithRetryDelay(delay time.Duration) RetryOption {
	return func(ro *RetryOptions) {
		ro.Delay = delay
	}
}

func WithRetryMethods(methods ...string) RetryOption {
	return func(ro *RetryOptions) {
		if ro.Methods == nil {
			ro.Methods = make(map[string]bool)
		}
		for _, method := range methods {
			ro.Methods[method] = true
		}
	}
}

func WithRetryAllMethods(retryAllMethods bool) RetryOption {
	return func(ro *RetryOptions) {
		ro.RetryAllMethods = retryAllMethods
	}
}
