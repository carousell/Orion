/*Package httptripper provides an implementation of http.RoundTripper that wraps zipking span info in request headers

Setup

for most cases wrapping the default http RoundTripper works
	tripper := httptripper.WrapTripper(http.DefaultTransport)
	http.DefaultTransport = tripper
doing this at the start of your program will make sure zipkin spans are appended for all outgoing http requests

Note: If you are using a custom tripper, then just wrap your custom tripper using httptripper.WrapTripper

How To Use

Make sure you add context info in the http.Request

	httpReq, err := http.NewRequest("GET", url, nil)
	httpReq = httpReq.WithContext(ctx)

*/
package httptripper

import (
	"net/http"

	"github.com/carousell/Orion/utils/spanutils"
	opentracing "github.com/opentracing/opentracing-go"
)

type tripper struct {
	transport     http.RoundTripper
	enableHystrix bool
}

func (t *tripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if opentracing.SpanFromContext(req.Context()) != nil {
		sp, _ := spanutils.NewHTTPExternalSpan(req.Context(), req.Host, req.URL.Path, req.Header)
		defer sp.Finish()
		resp, err := t.doRoundTrip(req)
		if err != nil {
			sp.SetTag("error", err.Error())
		}
		return resp, err
	}
	return t.doRoundTrip(req)
}

func (t *tripper) doRoundTrip(req *http.Request) (*http.Response, error) {
	return t.getTripper().RoundTrip(req)
}

func (t *tripper) getTripper() http.RoundTripper {
	if t.transport != nil {
		return t.transport
	}
	return http.DefaultTransport
}

//WrapTripper wraps the base tripper with zipkin info
func WrapTripper(base http.RoundTripper) http.RoundTripper {
	return &tripper{transport: base}
}
