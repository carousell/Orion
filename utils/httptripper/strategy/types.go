package strategy

import (
	"net/http"
	"time"
)

type Strategy interface {
	WaitDuration(retryCount, maxRetry int, req *http.Request, resp *http.Response, err error) time.Duration
}
