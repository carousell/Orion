package retry

import (
	"net/http"
	"time"

	"github.com/carousell/Orion/utils/httptripper/strategy"
)

//Retriable is the interface implemented by a retrier
type Retriable interface {
	//ShouldRetry should return when the failure needs to be retried
	ShouldRetry(attempt int, req *http.Request, resp *http.Response, err error) bool
	//WaitDuration should return the duration to wait before making next call
	WaitDuration(attempt int, req *http.Request, resp *http.Response, err error) time.Duration
}

//OptionsData stores all options used across retry
type OptionsData struct {
	MaxRetry        int
	Methods         map[string]bool
	RetryAllMethods bool
	Strategy        strategy.Strategy
}

//Option is the interface for defining Options
type Option func(*OptionsData)
