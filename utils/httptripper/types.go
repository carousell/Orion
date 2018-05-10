package httptripper

import (
	"net/http"
	"time"
)

type Retriable interface {
	ShouldRetry(retryConut int, req *http.Request, resp *http.Response, err error) bool
	WaitDuration(retryConut int, req *http.Request, resp *http.Response, err error) time.Duration
}

type RetryOptions struct {
	MaxRetry        int
	Delay           time.Duration
	Methods         map[string]bool
	RetryAllMethods bool
}

type RetryOption func(*RetryOptions)
