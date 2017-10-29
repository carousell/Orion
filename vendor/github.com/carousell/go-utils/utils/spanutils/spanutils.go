package spanutils

import (
	"context"
	"encoding/base64"
	"log"
	"strings"

	"github.com/carousell/go-utils/utils"
	newrelic "github.com/newrelic/go-agent"
	opentracing "github.com/opentracing/opentracing-go"
	otext "github.com/opentracing/opentracing-go/ext"
	"google.golang.org/grpc/metadata"
)

type TracingSpan interface {
	End()
	Finish()
	SetTag(key string, value interface{})
	SetQuery(query string)
}

type tracingSpan struct {
	openSpan        opentracing.Span
	datastore       bool
	external        bool
	dataSegment     newrelic.DatastoreSegment
	externalSegment newrelic.ExternalSegment
	segment         newrelic.Segment
}

func (span *tracingSpan) End() {
	if span.datastore {
		span.dataSegment.End()
	} else if span.external {
		span.externalSegment.End()
	} else {
		span.segment.End()
	}
	span.openSpan.Finish()
}

func (span *tracingSpan) Finish() {
	span.End()
}

func (span *tracingSpan) SetTag(key string, value interface{}) {
	span.openSpan.SetTag(key, value)
}

func (span *tracingSpan) SetQuery(query string) {
	span.openSpan.SetTag("query", query)
	if span.datastore {
		span.dataSegment.ParameterizedQuery = query
	}
}

func NewInternalSpan(ctx context.Context, name string) (TracingSpan, context.Context) {
	zip, ctx := opentracing.StartSpanFromContext(ctx, name)
	txn := utils.GetNewRelicTransactionFromContext(ctx)
	seg := newrelic.Segment{
		StartTime: newrelic.StartSegmentNow(txn),
		Name:      name,
	}
	return &tracingSpan{
		openSpan: zip,
		segment:  seg,
	}, ctx
}

func NewDatastoreSpan(ctx context.Context, name string, datastore string) (TracingSpan, context.Context) {
	if !strings.HasPrefix(name, datastore) {
		name = datastore + name
	}
	zip, ctx := opentracing.StartSpanFromContext(ctx, name)
	zip.SetTag("store", datastore)
	txn := utils.GetNewRelicTransactionFromContext(ctx)
	seg := newrelic.DatastoreSegment{
		StartTime: newrelic.StartSegmentNow(txn),
		Product:   newrelic.DatastoreProduct(datastore),
		Operation: name,
	}
	return &tracingSpan{
		openSpan:    zip,
		dataSegment: seg,
		datastore:   true,
	}, ctx
}

func NewExternalSpan(ctx context.Context, name string, url string) (TracingSpan, context.Context) {
	ctx, zip := ClientSpan(name, ctx)
	zip.SetTag("url", url)
	txn := utils.GetNewRelicTransactionFromContext(ctx)
	seg := newrelic.ExternalSegment{
		StartTime: newrelic.StartSegmentNow(txn),
		URL:       url,
	}
	return &tracingSpan{
		openSpan:        zip,
		externalSegment: seg,
		external:        true,
	}, ctx
}

// A type that conforms to opentracing.TextMapReader and
// opentracing.TextMapWriter.
type metadataReaderWriter struct {
	*metadata.MD
}

func (w metadataReaderWriter) Set(key, val string) {
	key = strings.ToLower(key)
	if strings.HasSuffix(key, "-bin") {
		val = string(base64.StdEncoding.EncodeToString([]byte(val)))
	}
	(*w.MD)[key] = append((*w.MD)[key], val)
}

func (w metadataReaderWriter) ForeachKey(handler func(key, val string) error) error {
	for k, vals := range *w.MD {
		for _, v := range vals {
			if err := handler(k, v); err != nil {
				return err
			}
		}
	}
	return nil
}

func ClientSpan(operationName string, ctx context.Context) (context.Context, opentracing.Span) {
	tracer := opentracing.GlobalTracer()
	var clientSpan opentracing.Span
	if parentSpan := opentracing.SpanFromContext(ctx); parentSpan != nil {
		clientSpan = tracer.StartSpan(
			operationName,
			opentracing.ChildOf(parentSpan.Context()),
		)
	} else {
		clientSpan = tracer.StartSpan(operationName)
	}
	otext.SpanKindRPCClient.Set(clientSpan)
	ctx = opentracing.ContextWithSpan(ctx, clientSpan)
	return ctx, clientSpan
}

func GRPCTracingSpan(operationName string, ctx context.Context) context.Context {
	tracer := opentracing.GlobalTracer()
	// Retrieve gRPC metadata.
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}
	if span := opentracing.SpanFromContext(ctx); span != nil {
		// There's nothing we can do with an error here.
		if err := tracer.Inject(span.Context(), opentracing.TextMap, metadataReaderWriter{&md}); err != nil {
			log.Println("err", err)
		}
	}

	var span opentracing.Span
	wireContext, err := tracer.Extract(opentracing.TextMap, metadataReaderWriter{&md})
	if err != nil && err != opentracing.ErrSpanContextNotFound {
		log.Println("err", err)
	}
	span = tracer.StartSpan(operationName, otext.RPCServerOption(wireContext))
	ctx = opentracing.ContextWithSpan(ctx, span)
	ctx = metadata.NewOutgoingContext(ctx, md)
	return ctx
}
