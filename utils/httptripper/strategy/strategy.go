/*
Package strategy provides strategies for use with retry
*/
package strategy

import (
	"math"
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
	if retryCount <= 0 {
		retryCount = 1
	}
	factor := int(math.Pow(2, float64(retryCount))) - 1
	return time.Duration(factor) * d.duration
}

//DefaultStrategy provides implementation for Fixed duration wait
func DefaultStrategy(duration time.Duration) Strategy {
	return &defaultStrategy{
		duration:    duration,
		exponential: false,
	}
}

//ExponentialStrategy provided implementation for exponentialy (in powers of 2) growing wait duration
func ExponentialStrategy(duration time.Duration) Strategy {
	return &defaultStrategy{
		duration:    duration,
		exponential: true,
	}
}
