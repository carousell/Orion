package spanutils

import (
	"testing"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

type testSpan struct {
	// spanContext testSpanContext
	// OperationName string
	// StartTime     time.Time
	// Tags          map[string]interface{}
	tags map[string]interface{}
}

func (n *testSpan) Context() opentracing.SpanContext { return nil }
func (n *testSpan) SetTag(key string, value interface{}) opentracing.Span {
	if n.tags == nil {
		n.tags = make(map[string]interface{})
	}
	n.tags[key] = value
	return n
}
func (n *testSpan) Finish()                                                {}
func (n *testSpan) FinishWithOptions(opts opentracing.FinishOptions)       {}
func (n *testSpan) LogFields(fields ...log.Field)                          {}
func (n *testSpan) LogKV(kvs ...interface{})                               {}
func (n *testSpan) SetOperationName(operationName string) opentracing.Span { return n }
func (n *testSpan) Tracer() opentracing.Tracer                             { return nil }
func (n *testSpan) SetBaggageItem(key, val string) opentracing.Span        { return n }
func (n *testSpan) BaggageItem(key string) string                          { return "" }
func (n *testSpan) LogEvent(event string)                                  {}
func (n *testSpan) LogEventWithPayload(event string, payload interface{})  {}
func (n *testSpan) Log(data opentracing.LogData)                           {}

func TestTracingSpan_SetTag(t *testing.T) {
	var tests = []struct {
		name  string
		given string
	}{
		{
			"tag is set on embedded opentracing.Span",
			"value_1",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			span := &testSpan{}
			tracingSpan := &tracingSpan{openSpan: span}
			tracingSpan.SetTag("key", tt.given)
			assertTagSet(t, span, "key", tt.given, tt.given)
		})
	}

	t.Run("no panic when called against nil span", func(t *testing.T) {
		var ts *tracingSpan
		ts.SetTag("k", "v")
	})
}

func TestTracingSpan_SetQuery(t *testing.T) {
	var tests = []struct {
		name  string
		given string
	}{
		{
			`value is set with tag="query"`,
			"SELECT * FROM tbl",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			span := &testSpan{}
			tracingSpan := &tracingSpan{openSpan: span}
			tracingSpan.SetQuery(tt.given)

			assertTagSet(t, span, "query", tt.given, tt.given)
		})
	}
	t.Run("no panic when called against nil span", func(t *testing.T) {
		var ts *tracingSpan
		ts.SetQuery("v")
	})
}

func assertTagSet(t *testing.T, span *testSpan, key, givenValue, expectedValue string) {
	t.Helper()
	if setValue, ok := span.tags[key]; ok {
		if v, ok := setValue.(string); ok {
			if v != expectedValue {
				t.Errorf("(key=%+v, value=%+v): expected value %+v to be set, got %+v", key, givenValue, expectedValue, v)
			}
		} else {
			t.Errorf("key=%+v: set value %+v is not a string ", key, v)
		}
	} else {
		t.Errorf("key=%+v: value not set for key", key)
	}
}
