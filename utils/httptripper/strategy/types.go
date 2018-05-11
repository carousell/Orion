package strategy

import (
	"net/http"
	"time"
)

//Strategy is the interface requirement for any strategy implementation
type Strategy interface {
	WaitDuration(retryCount, maxRetry int, req *http.Request, resp *http.Response, err error) time.Duration
}
