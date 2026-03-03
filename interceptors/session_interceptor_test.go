package interceptors

import (
	"context"
	"sync"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type mockProducer struct {
	mu        sync.Mutex
	published []publishedEvent
	notify    chan struct{}
}

type publishedEvent struct {
	topic string
	event interface{}
}

func (m *mockProducer) PublishAsync(topic string, event interface{}) error {
	m.mu.Lock()
	m.published = append(m.published, publishedEvent{topic: topic, event: event})
	m.mu.Unlock()
	if m.notify != nil {
		select {
		case m.notify <- struct{}{}:
		default:
		}
	}
	return nil
}

func newMockProducer() *mockProducer {
	return &mockProducer{notify: make(chan struct{}, 1)}
}

func (m *mockProducer) waitForEvent(t *testing.T, timeout time.Duration) publishedEvent {
	t.Helper()
	select {
	case <-m.notify:
	case <-time.After(timeout):
		t.Fatal("timed out waiting for published event")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.published[len(m.published)-1]
}

func okHandler(ctx context.Context, req interface{}) (interface{}, error) {
	return "ok", nil
}

func info(method string) *grpc.UnaryServerInfo {
	return &grpc.UnaryServerInfo{FullMethod: method}
}

func TestSessionActivityInterceptor_NoMetadata_PassesThrough(t *testing.T) {
	p := newMockProducer()
	interceptor := SessionActivityInterceptor("svc", p, DefaultSessionActivityTopic)
	resp, err := interceptor(context.Background(), "req", info("/svc/Method"), okHandler)
	if err != nil || resp != "ok" {
		t.Fatalf("unexpected resp=%v err=%v", resp, err)
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.published) != 0 {
		t.Error("should not publish without incoming metadata")
	}
}

func TestSessionActivityInterceptor_NoSessionContext_PassesThrough(t *testing.T) {
	p := newMockProducer()
	interceptor := SessionActivityInterceptor("svc", p, DefaultSessionActivityTopic)
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("other-header", "value"))
	resp, err := interceptor(ctx, "req", info("/svc/Method"), okHandler)
	if err != nil || resp != "ok" {
		t.Fatalf("unexpected resp=%v err=%v", resp, err)
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.published) != 0 {
		t.Error("should not publish when x-session-context is absent")
	}
}

func TestSessionActivityInterceptor_SessionContextPresent_Publishes(t *testing.T) {
	p := newMockProducer()
	encoded := "dGVzdC1zZXNzaW9u"
	interceptor := SessionActivityInterceptor("svc", p, "custom-topic")
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-session-context", encoded))

	resp, err := interceptor(ctx, "req", info("/svc/DoThing"), okHandler)
	if err != nil || resp != "ok" {
		t.Fatalf("unexpected resp=%v err=%v", resp, err)
	}

	got := p.waitForEvent(t, 2*time.Second)
	e := got.event.(SessionActivityEvent)
	if e.EncodedSessionContext != encoded {
		t.Errorf("EncodedSessionContext want %q got %q", encoded, e.EncodedSessionContext)
	}
	if e.Service != "svc" {
		t.Errorf("Service want svc got %s", e.Service)
	}
	if e.Action != "/svc/DoThing" {
		t.Errorf("Action want /svc/DoThing got %s", e.Action)
	}
	if got.topic != "custom-topic" {
		t.Errorf("topic want custom-topic got %s", got.topic)
	}
}

func TestSessionActivityInterceptor_NilProducer_DoesNotPanic(t *testing.T) {
	interceptor := SessionActivityInterceptor("svc", nil, DefaultSessionActivityTopic)
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-session-context", "dGVzdA=="))
	resp, err := interceptor(ctx, "req", info("/svc/Method"), okHandler)
	if err != nil || resp != "ok" {
		t.Fatalf("unexpected resp=%v err=%v", resp, err)
	}
}

func TestGlobalSessionActivityInterceptor_UsesGlobals(t *testing.T) {
	p := newMockProducer()
	SetGlobalSessionActivityProducer(p)
	SetGlobalSessionServiceName("global-svc")
	SetGlobalSessionActivityTopic("global-topic")
	t.Cleanup(func() {
		GlobalSessionActivityProducer = nil
		GlobalSessionServiceName = ""
		GlobalSessionActivityTopic = DefaultSessionActivityTopic
	})

	interceptor := GlobalSessionActivityInterceptor()
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-session-context", "dGVzdA=="))
	resp, err := interceptor(ctx, "req", info("/svc/Method"), okHandler)
	if err != nil || resp != "ok" {
		t.Fatalf("unexpected resp=%v err=%v", resp, err)
	}
	got := p.waitForEvent(t, 2*time.Second)
	if got.topic != "global-topic" {
		t.Errorf("topic want global-topic got %s", got.topic)
	}
}
