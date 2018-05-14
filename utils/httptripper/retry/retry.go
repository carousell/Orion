/*
Package retry provides an implementation for retrying http requests with multiple wait strategies
*/
package retry

import (
	"net/http"
	"time"

	"github.com/carousell/Orion/utils/httptripper/strategy"
)

var (
	defaultDelay   = time.Millisecond * 100
	defaultOptions = []Option{
		WithMaxRetry(3),
		WithRetryMethods(http.MethodGet, http.MethodOptions, http.MethodHead),
		WithRetryAllMethods(false),
		WithStrategy(strategy.DefaultStrategy(defaultDelay)),
	}
)

type defaultRetry struct {
	option *OptionsData
}

func (d *defaultRetry) ShouldRetry(retryConut int, req *http.Request, resp *http.Response, err error) bool {
	if resp != nil {
		// dont retry for anything less than 5XX
		if resp.StatusCode < 500 {
			return false
		}
	}
	if retryConut < d.option.MaxRetry && req != nil {
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

//NewRetry creates a new retry strategy
func NewRetry(options ...Option) Retriable {
	r := &defaultRetry{
		option: &OptionsData{},
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

//WithMaxRetry set the max number of times a request is tried
func WithMaxRetry(max int) Option {
	return func(ro *OptionsData) {
		ro.MaxRetry = max
	}
}

//WithRetryMethods specifies the methods that can be retried
func WithRetryMethods(methods ...string) Option {
	return func(ro *OptionsData) {
		if ro.Methods == nil {
			ro.Methods = make(map[string]bool)
		}
		for _, method := range methods {
			ro.Methods[method] = true
		}
	}
}

//WithRetryAllMethods sets retry on all HTTP methods, overrides WithRetryMethods
func WithRetryAllMethods(retryAllMethods bool) Option {
	return func(ro *OptionsData) {
		ro.RetryAllMethods = retryAllMethods
	}
}

//WithStrategy defines the backoff strategy to be used
func WithStrategy(s strategy.Strategy) Option {
	return func(ro *OptionsData) {
		ro.Strategy = s
	}
}
