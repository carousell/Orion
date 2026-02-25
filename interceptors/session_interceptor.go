package interceptors

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/carousell/Orion/utils/log"
)

// DefaultSessionActivityTopic is the default Kafka topic for session activity events.
const DefaultSessionActivityTopic = "session-activities"

// SessionActivityEvent is the event published to Kafka for AuthSvc consumption.
// EncodedSessionContext is the raw base64-encoded session context (x-session-context);
type SessionActivityEvent struct {
	EncodedSessionContext string `json:"encoded_session_context"`
	Service               string `json:"service"`
	Action                string `json:"action"`
	Status                string `json:"status"`
	DurationMs            int64  `json:"duration_ms"`
	Timestamp             int64  `json:"timestamp"`
}

// SessionActivityProducer is the interface for publishing session activity events (testable).
type SessionActivityProducer interface {
	PublishAsync(topic string, event interface{}) error
}

// GlobalSessionActivityProducer is the singleton producer used by GlobalSessionActivityInterceptor.
var GlobalSessionActivityProducer SessionActivityProducer

// GlobalSessionServiceName is the service name used by GlobalSessionActivityInterceptor.
var GlobalSessionServiceName string

// SetGlobalSessionActivityProducer sets the global producer (used by SessionInitializer).
func SetGlobalSessionActivityProducer(p SessionActivityProducer) {
	GlobalSessionActivityProducer = p
}

// SetGlobalSessionServiceName sets the global service name (used by SessionInitializer).
func SetGlobalSessionServiceName(name string) {
	GlobalSessionServiceName = name
}

// SessionActivityInterceptor runs only when a service opts in via GetInterceptors().
// It reads x-session-context from incoming metadata; if present and x-session-tracking is "true",
// async-publishes a session activity event (with encoded_session_context as-is) to Kafka. AuthSvc decodes on consume.
// Not added to DefaultInterceptors so only services that explicitly add it are affected.
func SessionActivityInterceptor(serviceName string, producer SessionActivityProducer) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return handler(ctx, req)
		}
		encoded := getMetadataValue(md, "x-session-context")
		if encoded == "" {
			log.Info(ctx, "Session context is missing; skipping further")
			return handler(ctx, req)
		}
		trackingHeader := getMetadataValue(md, "x-session-tracking")
		shouldTrack := trackingHeader == "true"
		if shouldTrack {
			ctx = metadata.AppendToOutgoingContext(ctx, "x-session-tracking", "true")
		}
		startTime := time.Now()
		resp, handlerErr := handler(ctx, req)
		duration := time.Since(startTime)
		if !shouldTrack || producer == nil {
			return resp, handlerErr
		}
		statusCode := "OK"
		if handlerErr != nil {
			statusCode = status.Convert(handlerErr).Code().String()
		}
		event := SessionActivityEvent{
			EncodedSessionContext: encoded,
			Service:               serviceName,
			Action:                info.FullMethod,
			Status:                statusCode,
			DurationMs:            duration.Milliseconds(),
			Timestamp:             time.Now().Unix(),
		}
		go func() {
			if err := producer.PublishAsync(DefaultSessionActivityTopic, event); err != nil {
				log.Error(context.Background(), "session_activity_publish_failed", "error", err, "service", serviceName, "action", info.FullMethod)
			}
		}()
		return resp, handlerErr
	}
}

// GlobalSessionActivityInterceptor returns an interceptor that uses GlobalSessionActivityProducer and GlobalSessionServiceName.
// Services that opt in (e.g. UserSvc) add this to their GetInterceptors() and register SessionInitializer in main.
func GlobalSessionActivityInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		serviceName := GlobalSessionServiceName
		if serviceName == "" {
			serviceName = "unknown-service"
		}
		return SessionActivityInterceptor(serviceName, GlobalSessionActivityProducer)(ctx, req, info, handler)
	}
}

func getMetadataValue(md metadata.MD, key string) string {
	values := md.Get(key)
	if len(values) > 0 {
		return values[0]
	}
	return ""
}
