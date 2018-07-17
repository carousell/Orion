package worker

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/carousell/Orion/utils/log"
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
		log.Info(ctx, "trace", "not found")
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
		log.Info(context.Background(), "schedule", "scheduling error", "error", err)
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
		log.Error(context.Background(), "worker", "can not deserialize work", "error", err)
		return nil
	}
	return wi
}
