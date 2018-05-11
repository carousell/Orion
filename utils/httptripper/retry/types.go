package retry

import (
	"net/http"
	"time"

	"github.com/carousell/Orion/utils/httptripper/strategy"
)

type Retriable interface {
	ShouldRetry(retryConut int, req *http.Request, resp *http.Response, err error) bool
	WaitDuration(retryConut int, req *http.Request, resp *http.Response, err error) time.Duration
}

type RetryOptions struct {
	MaxRetry        int
	Methods         map[string]bool
	RetryAllMethods bool
	Strategy        strategy.Strategy
}

type RetryOption func(*RetryOptions)
