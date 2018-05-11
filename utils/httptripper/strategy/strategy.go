package strategy

import (
	"net/http"
	"time"
)

type defaultStrategy struct {
	duration    time.Duration
	exponential bool
}

func (d *defaultStrategy) WaitDuration(retryCount int, maxRetry int, req *http.Request, resp *http.Response, err error) time.Duration {
	if !d.exponential {
		return d.duration
	}
	if retryCount == 0 {
		retryCount = 1
	}
	factor := 2 ^ (retryCount - 1)
	return time.Duration(factor) * d.duration
}

func DefaultStrategy(duration time.Duration) Strategy {
	return &defaultStrategy{
		duration:    duration,
		exponential: false,
	}
}

func ExponentialStrategy(duration time.Duration) Strategy {
	return &defaultStrategy{
		duration:    duration,
		exponential: true,
	}
}
