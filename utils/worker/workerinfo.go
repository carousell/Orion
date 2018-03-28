package worker

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/satori/go.uuid"
)

type workerInfo struct {
	Name    string              `json:"name"`
	ID      string              `json:"id"`
	Trace   map[string][]string `json:"trace"`
	Payload string              `json:"payload"`
}

func (w *workerInfo) MarshalTraceInfo(ctx context.Context) {
	if w.Trace == nil {
		w.Trace = http.Header{}
	}
	if sp := opentracing.SpanFromContext(ctx); sp != nil {
		opentracing.GlobalTracer().Inject(
			sp.Context(),
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(w.Trace))
	} else {
		log.Println("Trace", "not found", ctx)
	}
}

func (w *workerInfo) UnmarshalTraceInfo(ctx context.Context) context.Context {
	wireContext, err := opentracing.GlobalTracer().Extract(
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(w.Trace))
	if err == nil {
		md := metautils.ExtractIncoming(ctx)
		opentracing.GlobalTracer().Inject(wireContext, opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(md))
		ctx = md.ToIncoming(ctx)
	}
	return ctx
}

func (w *workerInfo) String() string {
	data, err := json.Marshal(w)
	if err != nil {
		log.Println("schedule", "scheduling error", "error", err.Error())
		return ""
	}
	return string(data)
}

func newWorkerInfo(ctx context.Context, payload string) *workerInfo {
	uuidValue, _ := uuid.NewV4()
	wi := &workerInfo{
		ID:      uuidValue.String(),
		Payload: payload,
	}
	wi.MarshalTraceInfo(ctx)
	return wi
}

func unmarshalWorkerInfo(payload string) *workerInfo {
	wi := new(workerInfo)
	err := json.Unmarshal([]byte(payload), wi)
	if err != nil {
		log.Println("worker", "can not deserialize work", "error", err.Error())
		return nil
	}
	return wi
}
