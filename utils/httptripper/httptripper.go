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
	"errors"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/carousell/Orion/utils/httptripper/retry"
	"github.com/carousell/Orion/utils/spanutils"
)

var (
	defaultOptions = []Option{
		WithBaseTripper(http.DefaultTransport),
		WithHystrix(true),
		WithRetrier(retry.NewRetry()),
	}
)

type tripper struct {
	option *OptionsData
}

func (t *tripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if req == nil {
		return nil, errors.New("Nil request received")
	}
	attempt := 0
	var resp *http.Response
	var err error
	for attempt == 0 || t.getRetrier(req).ShouldRetry(attempt, req, resp, err) {
		// close body of previous response on retry
		if resp != nil {
			go resp.Body.Close()
		}
		if attempt != 0 {
			time.Sleep(t.getRetrier(req).WaitDuration(attempt, req, resp, err))
		}
		resp, err = t.doRoundTrip(req, attempt)
		attempt++
	}
	return resp, err
}

func (t *tripper) doRoundTrip(req *http.Request, retryConut int) (*http.Response, error) {
	traceName := GetRequestTraceName(req)
	if traceName == "" {
		traceName = req.Host
	}
	if !t.option.HystrixEnabled {
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
	if t.option.BaseTripper != nil {
		return t.option.BaseTripper
	}
	return http.DefaultTransport
}

func (t *tripper) getRetrier(req *http.Request) retry.Retriable {
	r := GetRequestRetrier(req)
	if r != nil {
		return r
	}
	if t.option.Retrier != nil {
		return t.option.Retrier
	}
	return retry.NewRetry()
}

//WrapTripper wraps the base tripper with zipkin info
func WrapTripper(base http.RoundTripper) http.RoundTripper {
	return NewTripper(WithBaseTripper(base))
}

//NewTripper returns a default tripper wrapped around http.DefaultTransport
func NewTripper(options ...Option) http.RoundTripper {
	t := &tripper{
		option: &OptionsData{},
	}
	for _, opt := range defaultOptions {
		opt(t.option)
	}
	for _, opt := range options {
		opt(t.option)
	}
	return t
}

//NewHTTPClient creates a new http.Client with default retry options and timeout
func NewHTTPClient(timeout time.Duration, options ...Option) *http.Client {
	if timeout == 0 {
		// never use a 0 timeout
		timeout = time.Second
	}
	return &http.Client{
		Transport: NewTripper(options...),
		Timeout:   timeout,
	}
}

//NewRequest extends http.NewRequest with context and trace name
func NewRequest(ctx context.Context, traceName, method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return req, err
	}
	return SetRequestTraceName(req.WithContext(ctx), traceName), nil
}

//NewRequestWithRetrier extends http.NewRequest with context, trace name and retrier
func NewRequestWithRetrier(ctx context.Context, traceName string, retrier retry.Retriable, method, url string, body io.Reader) (*http.Request, error) {
	req, err := NewRequest(ctx, traceName, method, url, body)
	if err != nil {
		return req, err
	}
	return SetRequestRetrier(req, retrier), nil
}

//SetRequestTraceName stores a trace name in a HTTP request
func SetRequestTraceName(req *http.Request, traceName string) *http.Request {
	ctx := req.Context()
	ctx = context.WithValue(ctx, traceID, traceName)
	return req.WithContext(ctx)
}

//GetRequestTraceName fetches a trace name from HTTP request
func GetRequestTraceName(req *http.Request) string {
	ctx := req.Context()
	if value := ctx.Value(traceID); value != nil {
		if traceName, ok := value.(string); ok {
			return traceName
		}
	}
	return ""
}

// SetRequestRetrier sets the retrier to be used with this request
func SetRequestRetrier(req *http.Request, retrier retry.Retriable) *http.Request {
	ctx := req.Context()
	ctx = context.WithValue(ctx, retrierKey, retrier)
	return req.WithContext(ctx)
}

//GetRequestRetrier fetches retrier to be used with this request
func GetRequestRetrier(req *http.Request) retry.Retriable {
	ctx := req.Context()
	if value := ctx.Value(retrierKey); value != nil {
		if retrier, ok := value.(retry.Retriable); ok {
			return retrier
		}
	}
	return nil
}

//WithBaseTripper updates the tripper to use the provided http.RoundTripper
func WithBaseTripper(base http.RoundTripper) Option {
	return func(o *OptionsData) {
		o.BaseTripper = base
	}
}

//WithRetrier updates the tripper to use the provided retry.Retriable
func WithRetrier(retrier retry.Retriable) Option {
	return func(o *OptionsData) {
		o.Retrier = retrier
	}
}

//WithHystrix enables/disables use of hystrix
func WithHystrix(enabled bool) Option {
	return func(o *OptionsData) {
		o.HystrixEnabled = enabled
	}
}
