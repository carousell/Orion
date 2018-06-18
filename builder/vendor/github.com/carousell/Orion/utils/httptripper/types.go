package httptripper

import (
	"net/http"

	"github.com/carousell/Orion/utils/httptripper/retry"
)

type key string

const (
	traceID    key = "HTTPRequestTracingName"
	retrierKey key = "HTTPRequestRetrier"
)

//OptionsData is the data polulated by the options
type OptionsData struct {
	BaseTripper    http.RoundTripper
	HystrixEnabled bool
	Retrier        retry.Retriable
}

// Option defines an options for Tripper
type Option func(*OptionsData)
