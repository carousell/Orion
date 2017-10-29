package ServiceName

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/carousell/Orion/example/ServiceName/ServiceName_proto"
	"github.com/carousell/go-utils/utils"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/tracing/opentracing"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	stdopentracing "github.com/opentracing/opentracing-go"
)

const TracerKey = "CF-RAY"

func encodeAndSendJSON(w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(response)
}

func EncodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(response)
}

func DecodeEchoRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	protoReq := new(ServiceName_proto.EchoRequest)
	protoReq.Msg = r.URL.Query().Get("msg")
	return protoReq, nil
}

func DecodeUppercaseRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	protoReq := new(ServiceName_proto.UppercaseRequest)
	protoReq.Msg = r.URL.Query().Get("msg")
	return protoReq, nil
}

func EncodeEchoReponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	if msg, ok := response.(string); ok {
		resp := new(ServiceName_proto.EchoResponse)
		resp.Msg = msg
		return encodeAndSendJSON(w, resp)
	} else {
		return utils.NewCustomError("could not encode response", 500, response)
	}
}

// error encoder for HTTP request
func ErrorEncoder(ctx context.Context, err error, w http.ResponseWriter) {
	var resp interface{}
	// check for custom error
	if e, ok := err.(utils.CustomError); ok {
		if e.StatusCode != 0 {
			w.WriteHeader(e.StatusCode)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		if e.Payload != nil {
			resp = e.Payload
		}
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}

	// build a resp object
	encodeAndSendJSON(w, resp)
}

func newHTTPServer(name string, logger log.Logger, tracer stdopentracing.Tracer, e endpoint.Endpoint, dec httptransport.DecodeRequestFunc, enc httptransport.EncodeResponseFunc) *httptransport.Server {
	logger = log.With(logger, "Endpoint", name)
	options := []httptransport.ServerOption{
		httptransport.ServerErrorEncoder(ErrorEncoder),
		httptransport.ServerErrorLogger(logger),
		httptransport.ServerBefore(opentracing.HTTPToContext(tracer, name, logger)),
	}
	return httptransport.NewServer(
		utils.EndpointLoggingMiddleware(logger)(e),
		utils.UpdateHTTPTracingSpanWithTag(dec, TracerKey),
		enc,
		options...,
	)
}

func MakeHTTPHandler(ctx context.Context, endpoints Endpoints, logger log.Logger, tracer stdopentracing.Tracer) http.Handler {
	r := mux.NewRouter()
	// write calls
	r.Methods("GET", "POST").Path("/echo/").Handler(
		newHTTPServer(
			"Echo",
			logger,
			tracer,
			endpoints.Echo,
			DecodeEchoRequest,
			EncodeEchoReponse,
		),
	)
	r.Methods("GET", "POST").Path("/uppercase/").Handler(
		newHTTPServer(
			"Uppercase",
			logger,
			tracer,
			endpoints.Uppercase,
			DecodeUppercaseRequest,
			EncodeResponse,
		),
	)
	return r
}
