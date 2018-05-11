package retry

import (
	"net/http"
	"time"

	"github.com/carousell/Orion/utils/httptripper/strategy"
)

var (
	defaultDelay   = time.Millisecond * 100
	defaultOptions = []RetryOption{
		WithMaxRetry(3),
		WithRetryMethods(http.MethodGet, http.MethodOptions, http.MethodHead),
		WithRetryAllMethods(false),
		WithStrategy(strategy.DefaultStrategy(defaultDelay)),
	}
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
	if d.option.Strategy == nil {
		return defaultDelay
	}
	return d.option.Strategy.WaitDuration(retryConut, d.option.MaxRetry, req, resp, err)
}

func NewRetry(options ...RetryOption) Retriable {
	r := &defaultRetry{
		option: &RetryOptions{},
	}
	// apply default
	for _, opt := range defaultOptions {
		opt(r.option)
	}
	// apply user provided
	for _, opt := range options {
		opt(r.option)
	}
	return r
}

func WithMaxRetry(max int) RetryOption {
	return func(ro *RetryOptions) {
		ro.MaxRetry = max
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

func WithStrategy(s strategy.Strategy) RetryOption {
	return func(ro *RetryOptions) {
		ro.Strategy = s
	}
}
