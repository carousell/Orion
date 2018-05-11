/*Package httptripper provides an implementation of http.RoundTripper that provides retries, popluates opentracing span info and hystrix circuit breaker.

Setup

for most cases using the http.Client provided by the package is sufficient
	client := httptripper.NewHTTPClient(time.Millisecond * 500)

Note: If you are using a custom http.Client, then just wrap your custom http.Client using httptripper.WrapTripper
	tripper := httptripper.WrapTripper(client.Transport)
	client.Transport = tripper

How To Use

Make sure you use httptripper.NewRequest to build http.Request, since http.NewRequest does not take context as parameter
	httpReq, err := httptripper.NewRequest(ctx, "TracingName", "GET", url, nil)

*/
package httptripper

import (
	"context"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/carousell/Orion/utils/httptripper/retry"
	"github.com/carousell/Orion/utils/spanutils"
)

const (
	traceID = "HTTPRequestTracingName"
)

type tripper struct {
	transport      http.RoundTripper
	retrier        retry.Retriable
	hystrixEnabled bool
}

func (t *tripper) RoundTrip(req *http.Request) (*http.Response, error) {
	retry := 0
	firstTry := true
	var resp *http.Response
	var err error
	for firstTry || t.getRetrier().ShouldRetry(retry, req, resp, err) {
		if !firstTry {
			time.Sleep(t.getRetrier().WaitDuration(retry, req, resp, err))
		}
		resp, err = t.doRoundTrip(req, retry)
		firstTry = false
		retry += 1
	}
	return resp, err
}

func (t *tripper) doRoundTrip(req *http.Request, retryConut int) (*http.Response, error) {
	traceName := GetRequestTraceName(req)
	if traceName == "" {
		traceName = req.Host
	}
	if !t.hystrixEnabled {
		// hystrix not enabled go ahead without it
		return t.makeRoundTrip(traceName, req, retryConut)
	}
	var resp *http.Response
	var err error
	err = hystrix.Do(
		traceName,
		func() error {
			resp, err = t.makeRoundTrip(traceName, req, retryConut)
			return err
		},
		nil,
	)
	return resp, err
}

func (t *tripper) makeRoundTrip(traceName string, req *http.Request, retryConut int) (*http.Response, error) {
	span, _ := spanutils.NewHTTPExternalSpan(req.Context(),
		traceName, req.URL.String(), req.Header)
	defer span.Finish()
	span.SetTag("attempt", retryConut)
	resp, err := t.getTripper().RoundTrip(req)
	if err != nil {
		span.SetTag("error", err.Error())
		if e, ok := err.(net.Error); ok && e.Timeout() {
			span.SetTag("timeout", true)
		}
	}
	if resp != nil {
		span.SetTag("statuscode", resp.StatusCode)
	}
	return resp, err
}

func (t *tripper) getTripper() http.RoundTripper {
	if t.transport != nil {
		return t.transport
	}
	return http.DefaultTransport
}

func (t *tripper) getRetrier() retry.Retriable {
	if t.retrier != nil {
		return t.retrier
	}
	return retry.NewRetry()
}

//WrapTripper wraps the base tripper with zipkin info
func WrapTripper(base http.RoundTripper) http.RoundTripper {
	return &tripper{
		transport:      base,
		retrier:        retry.NewRetry(),
		hystrixEnabled: false,
	}
}

func NewTripper() http.RoundTripper {
	return &tripper{
		retrier:        retry.NewRetry(),
		transport:      http.DefaultTransport,
		hystrixEnabled: true,
	}
}

func NewHTTPClient(timeout time.Duration) *http.Client {
	if timeout == 0 {
		// never use a 0 timeout
		timeout = time.Second
	}
	return &http.Client{
		Transport: NewTripper(),
		Timeout:   timeout,
	}
}

func NewRequest(ctx context.Context, traceName, method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return req, err
	}
	return SetRequestTraceName(req.WithContext(ctx), traceName), err
}

func SetRequestTraceName(req *http.Request, traceName string) *http.Request {
	ctx := req.Context()
	ctx = context.WithValue(ctx, traceID, traceName)
	return req.WithContext(ctx)
}

func GetRequestTraceName(req *http.Request) string {
	ctx := req.Context()
	if value := ctx.Value(traceID); value != nil {
		if traceName, ok := value.(string); ok {
			return traceName
		}
	}
	return ""
}
