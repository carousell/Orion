package interceptors

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
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

func (m *mockProducer) eventCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.published)
}

func okHandler(ctx context.Context, req interface{}) (interface{}, error) {
	return "ok", nil
}

func errorHandler(ctx context.Context, req interface{}) (interface{}, error) {
	return nil, status.Error(codes.PermissionDenied, "forbidden")
}

func info(method string) *grpc.UnaryServerInfo {
	return &grpc.UnaryServerInfo{FullMethod: method}
}

// ctxWithSessionTrack builds a context with x-session-track, x-session-context,
// and optionally x-session-operation in incoming metadata.
func ctxWithSessionTrack(sessionCtx, operation string) context.Context {
	pairs := []string{"x-session-track", "true", "x-session-context", sessionCtx}
	if operation != "" {
		pairs = append(pairs, "x-session-operation", operation)
	}
	return metadata.NewIncomingContext(context.Background(), metadata.Pairs(pairs...))
}

// ── Tests for x-session-track guard ──

func TestSessionActivityInterceptor_NoMetadata_PassesThrough(t *testing.T) {
	p := newMockProducer()
	interceptor := SessionActivityInterceptor("svc", p, DefaultSessionActivityTopic)
	resp, err := interceptor(context.Background(), "req", info("/svc/Method"), okHandler)
	if err != nil || resp != "ok" {
		t.Fatalf("unexpected resp=%v err=%v", resp, err)
	}
	if p.eventCount() != 0 {
		t.Error("should not publish without incoming metadata")
	}
}

func TestSessionActivityInterceptor_NoSessionTrack_PassesThrough(t *testing.T) {
	p := newMockProducer()
	interceptor := SessionActivityInterceptor("svc", p, DefaultSessionActivityTopic)
	// Has x-session-context but NO x-session-track → should NOT publish
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		"x-session-context", "dGVzdC1zZXNzaW9u",
	))
	resp, err := interceptor(ctx, "req", info("/svc/Method"), okHandler)
	if err != nil || resp != "ok" {
		t.Fatalf("unexpected resp=%v err=%v", resp, err)
	}
	if p.eventCount() != 0 {
		t.Error("should not publish when x-session-track is absent even if x-session-context is present")
	}
}

func TestSessionActivityInterceptor_SessionTrackFalse_PassesThrough(t *testing.T) {
	p := newMockProducer()
	interceptor := SessionActivityInterceptor("svc", p, DefaultSessionActivityTopic)
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		"x-session-track", "false",
		"x-session-context", "dGVzdA==",
	))
	resp, err := interceptor(ctx, "req", info("/svc/Method"), okHandler)
	if err != nil || resp != "ok" {
		t.Fatalf("unexpected resp=%v err=%v", resp, err)
	}
	if p.eventCount() != 0 {
		t.Error("should not publish when x-session-track is not 'true'")
	}
}

func TestSessionActivityInterceptor_NoSessionContext_PassesThrough(t *testing.T) {
	p := newMockProducer()
	interceptor := SessionActivityInterceptor("svc", p, DefaultSessionActivityTopic)
	// x-session-track is true but x-session-context is missing
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		"x-session-track", "true",
		"other-header", "value",
	))
	resp, err := interceptor(ctx, "req", info("/svc/Method"), okHandler)
	if err != nil || resp != "ok" {
		t.Fatalf("unexpected resp=%v err=%v", resp, err)
	}
	if p.eventCount() != 0 {
		t.Error("should not publish when x-session-context is absent")
	}
}

// ── Tests for event publishing ──

func TestSessionActivityInterceptor_SessionTrackAndContextPresent_Publishes(t *testing.T) {
	p := newMockProducer()
	encoded := "dGVzdC1zZXNzaW9u"
	interceptor := SessionActivityInterceptor("svc", p, "custom-topic")
	ctx := ctxWithSessionTrack(encoded, "")

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
	// Without x-session-operation, Action falls back to FullMethod
	if e.Action != "/svc/DoThing" {
		t.Errorf("Action want /svc/DoThing got %s", e.Action)
	}
	if e.Status != "OK" {
		t.Errorf("Status want OK got %s", e.Status)
	}
	if got.topic != "custom-topic" {
		t.Errorf("topic want custom-topic got %s", got.topic)
	}
}

func TestSessionActivityInterceptor_WithOperationLabel_UsesOperationAsAction(t *testing.T) {
	p := newMockProducer()
	interceptor := SessionActivityInterceptor("svc", p, DefaultSessionActivityTopic)
	ctx := ctxWithSessionTrack("dGVzdA==", "change_password")

	resp, err := interceptor(ctx, "req", info("/svc/ChangePassword"), okHandler)
	if err != nil || resp != "ok" {
		t.Fatalf("unexpected resp=%v err=%v", resp, err)
	}

	got := p.waitForEvent(t, 2*time.Second)
	e := got.event.(SessionActivityEvent)
	if e.Action != "change_password" {
		t.Errorf("Action want change_password got %s", e.Action)
	}
}

func TestSessionActivityInterceptor_HandlerError_CapturesGRPCStatus(t *testing.T) {
	p := newMockProducer()
	interceptor := SessionActivityInterceptor("svc", p, DefaultSessionActivityTopic)
	ctx := ctxWithSessionTrack("dGVzdA==", "update_profile")

	resp, err := interceptor(ctx, "req", info("/svc/UpdateProfile"), errorHandler)
	if resp != nil {
		t.Fatalf("expected nil resp, got %v", resp)
	}
	if err == nil {
		t.Fatal("expected handler error")
	}

	got := p.waitForEvent(t, 2*time.Second)
	e := got.event.(SessionActivityEvent)
	if e.Status != "PermissionDenied" {
		t.Errorf("Status want PermissionDenied got %s", e.Status)
	}
	if e.Action != "update_profile" {
		t.Errorf("Action want update_profile got %s", e.Action)
	}
}

// ── Tests for strip-on-consume ──

func TestSessionActivityInterceptor_StripsTrackHeader_FromIncomingContext(t *testing.T) {
	p := newMockProducer()
	interceptor := SessionActivityInterceptor("svc", p, DefaultSessionActivityTopic)
	ctx := ctxWithSessionTrack("dGVzdA==", "op_name")

	// Handler checks that x-session-track is stripped from incoming metadata
	verifyHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, fmt.Errorf("expected incoming metadata")
		}
		if v := md.Get("x-session-track"); len(v) > 0 {
			return nil, fmt.Errorf("x-session-track should be stripped, got %v", v)
		}
		if v := md.Get("x-session-operation"); len(v) > 0 {
			return nil, fmt.Errorf("x-session-operation should be stripped, got %v", v)
		}
		// x-session-context should still be present
		if v := md.Get("x-session-context"); len(v) == 0 || v[0] != "dGVzdA==" {
			return nil, fmt.Errorf("x-session-context should be preserved, got %v", v)
		}
		return "verified", nil
	}

	resp, err := interceptor(ctx, "req", info("/svc/Method"), verifyHandler)
	if err != nil {
		t.Fatalf("handler verification failed: %v", err)
	}
	if resp != "verified" {
		t.Fatalf("unexpected resp=%v", resp)
	}
}

func TestSessionActivityInterceptor_StrippedHeadersNotRelayedByForwardMetadata(t *testing.T) {
	// Simulate what ForwardMetadataInterceptor does: read incoming and copy to outgoing.
	// After the session interceptor strips the headers, ForwardMetadataInterceptor should NOT
	// see them.
	p := newMockProducer()
	interceptor := SessionActivityInterceptor("svc", p, DefaultSessionActivityTopic)
	ctx := ctxWithSessionTrack("dGVzdA==", "some_op")

	verifyHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		// Simulate ForwardMetadataInterceptor behavior
		md, _ := metadata.FromIncomingContext(ctx)
		for key, values := range md {
			for _, value := range values {
				ctx = metadata.AppendToOutgoingContext(ctx, key, value)
			}
		}
		// Outgoing should NOT have x-session-track or x-session-operation
		outMD, _ := metadata.FromOutgoingContext(ctx)
		if v := outMD.Get("x-session-track"); len(v) > 0 {
			return nil, fmt.Errorf("ForwardMetadata should not relay x-session-track, got %v", v)
		}
		if v := outMD.Get("x-session-operation"); len(v) > 0 {
			return nil, fmt.Errorf("ForwardMetadata should not relay x-session-operation, got %v", v)
		}
		// x-session-context SHOULD be relayed (it's not stripped)
		if v := outMD.Get("x-session-context"); len(v) == 0 {
			return nil, fmt.Errorf("ForwardMetadata should relay x-session-context")
		}
		return "ok", nil
	}

	resp, err := interceptor(ctx, "req", info("/svc/Method"), verifyHandler)
	if err != nil {
		t.Fatalf("verification failed: %v", err)
	}
	if resp != "ok" {
		t.Fatalf("unexpected resp=%v", resp)
	}
}

// ── Tests for nil producer ──

func TestSessionActivityInterceptor_NilProducer_DoesNotPanic(t *testing.T) {
	// Reset the once so this test can trigger it
	warnNilProducerOnce = sync.Once{}
	interceptor := SessionActivityInterceptor("svc", nil, DefaultSessionActivityTopic)
	ctx := ctxWithSessionTrack("dGVzdA==", "")
	resp, err := interceptor(ctx, "req", info("/svc/Method"), okHandler)
	if err != nil || resp != "ok" {
		t.Fatalf("unexpected resp=%v err=%v", resp, err)
	}
}

// ── Tests for GlobalSessionActivityInterceptor ──

func TestGlobalSessionActivityInterceptor_UsesGlobals(t *testing.T) {
	p := newMockProducer()
	SetGlobalSessionActivityProducer(p)
	SetGlobalSessionServiceName("global-svc")
	SetGlobalSessionActivityTopic("global-topic")
	t.Cleanup(func() {
		GlobalSessionActivityProducer = nil
		GlobalSessionServiceName = ""
		GlobalSessionActivityTopic = DefaultSessionActivityTopic
		warnNilProducerOnce = sync.Once{}
	})

	interceptor := GlobalSessionActivityInterceptor()
	ctx := ctxWithSessionTrack("dGVzdA==", "")
	resp, err := interceptor(ctx, "req", info("/svc/Method"), okHandler)
	if err != nil || resp != "ok" {
		t.Fatalf("unexpected resp=%v err=%v", resp, err)
	}
	got := p.waitForEvent(t, 2*time.Second)
	if got.topic != "global-topic" {
		t.Errorf("topic want global-topic got %s", got.topic)
	}
	e := got.event.(SessionActivityEvent)
	if e.Service != "global-svc" {
		t.Errorf("Service want global-svc got %s", e.Service)
	}
}

func TestGlobalSessionActivityInterceptor_SkipsWithoutSessionTrack(t *testing.T) {
	p := newMockProducer()
	SetGlobalSessionActivityProducer(p)
	SetGlobalSessionServiceName("global-svc")
	t.Cleanup(func() {
		GlobalSessionActivityProducer = nil
		GlobalSessionServiceName = ""
		GlobalSessionActivityTopic = DefaultSessionActivityTopic
	})

	interceptor := GlobalSessionActivityInterceptor()
	// Only x-session-context, no x-session-track
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		"x-session-context", "dGVzdA==",
	))
	resp, err := interceptor(ctx, "req", info("/svc/Method"), okHandler)
	if err != nil || resp != "ok" {
		t.Fatalf("unexpected resp=%v err=%v", resp, err)
	}
	if p.eventCount() != 0 {
		t.Error("should not publish when x-session-track is absent")
	}
}
