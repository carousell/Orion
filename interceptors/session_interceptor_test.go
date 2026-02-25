package interceptors

import (
	"context"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestSessionActivityInterceptor_NoMetadata_PassesThrough(t *testing.T) {
	handlerCalled := false
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		handlerCalled = true
		return "ok", nil
	}
	interceptor := SessionActivityInterceptor("test-svc", nil)
	ctx := context.Background()
	resp, err := interceptor(ctx, "req", &grpc.UnaryServerInfo{FullMethod: "/test/Method"}, handler)
	if err != nil {
		t.Fatal(err)
	}
	if resp != "ok" {
		t.Errorf("resp want ok got %v", resp)
	}
	if !handlerCalled {
		t.Error("handler should have been called")
	}
}

func TestSessionActivityInterceptor_NoSessionContext_PassesThrough(t *testing.T) {
	handlerCalled := false
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		handlerCalled = true
		return "ok", nil
	}
	interceptor := SessionActivityInterceptor("test-svc", nil)
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("other", "value"))
	resp, err := interceptor(ctx, "req", &grpc.UnaryServerInfo{FullMethod: "/test/Method"}, handler)
	if err != nil {
		t.Fatal(err)
	}
	if resp != "ok" {
		t.Errorf("resp want ok got %v", resp)
	}
	if !handlerCalled {
		t.Error("handler should have been called")
	}
}

func TestSessionActivityInterceptor_WithSessionContext_NoTracking_HandlerCalled(t *testing.T) {
	handlerCalled := false
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		handlerCalled = true
		return "ok", nil
	}
	interceptor := SessionActivityInterceptor("test-svc", nil)
	// Encoded value is passed through as-is; no decoding in Orion. Use any non-empty string.
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-session-context", "dGVzdA=="))
	resp, err := interceptor(ctx, "req", &grpc.UnaryServerInfo{FullMethod: "/test/Method"}, handler)
	if err != nil {
		t.Fatal(err)
	}
	if resp != "ok" {
		t.Errorf("resp want ok got %v", resp)
	}
	if !handlerCalled {
		t.Error("handler should have been called")
	}
}

func TestSessionActivityInterceptor_WithTrackingAndProducer_PublishesAsync(t *testing.T) {
	done := make(chan SessionActivityEvent, 1)
	mockProducer := &mockSessionProducer{
		publish: func(topic string, event interface{}) error {
			if e, ok := event.(SessionActivityEvent); ok {
				select {
				case done <- e:
				default:
				}
			}
			return nil
		},
	}
	// No proto encoding in Orion; any non-empty encoded string is forwarded as-is to Kafka. AuthSvc decodes on consume.
	encoded := "dGVzdC1zZXNzaW9uLWNvbnRleHQ="
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "ok", nil
	}
	interceptor := SessionActivityInterceptor("test-svc", mockProducer)
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		"x-session-context", encoded,
		"x-session-tracking", "true",
	))
	resp, err := interceptor(ctx, "req", &grpc.UnaryServerInfo{FullMethod: "/user.UserService/UpdateProfile"}, handler)
	if err != nil {
		t.Fatal(err)
	}
	if resp != "ok" {
		t.Errorf("resp want ok got %v", resp)
	}
	e := <-done
	if e.Service != "test-svc" {
		t.Errorf("Service want test-svc got %s", e.Service)
	}
	if e.Action != "/user.UserService/UpdateProfile" {
		t.Errorf("Action want /user.UserService/UpdateProfile got %s", e.Action)
	}
	if e.EncodedSessionContext != encoded {
		t.Errorf("EncodedSessionContext want %q got %q", encoded, e.EncodedSessionContext)
	}
}

type mockSessionProducer struct {
	publish func(topic string, event interface{}) error
}

func (m *mockSessionProducer) PublishAsync(topic string, event interface{}) error {
	if m.publish != nil {
		return m.publish(topic, event)
	}
	return nil
}
