package strategy

import (
	"net/http"
	"time"
)

//Strategy is the interface requirement for any strategy implementation
type Strategy interface {
	//WaitDuration takes attempt, maxRetry and request/response paramaetrs as input and gives out a duration as response
	WaitDuration(attempt, maxRetry int, req *http.Request, resp *http.Response, err error) time.Duration
}
