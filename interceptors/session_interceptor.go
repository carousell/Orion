package interceptors

import (
	"context"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/carousell/Orion/utils/log"
)

// DefaultSessionActivityTopic is the default Kafka topic for session activity events.
const DefaultSessionActivityTopic = "session-activities"

// SessionActivityEvent is published to Kafka whenever x-session-context is present.
// EncodedSessionContext is the raw base64-encoded value forwarded as-is; consumers decode it.
type SessionActivityEvent struct {
	EncodedSessionContext string `json:"encoded_session_context"`
	Service               string `json:"service"`
	Action                string `json:"action"`
	Status                string `json:"status"`
	DurationMs            int64  `json:"duration_ms"`
	Timestamp             int64  `json:"timestamp"`
}

// SessionActivityProducer publishes session activity events to a Kafka topic.
type SessionActivityProducer interface {
	PublishAsync(topic string, event interface{}) error
}

// SessionActivityInterceptor is an optional gRPC server interceptor that publishes a
// SessionActivityEvent to the given Kafka topic whenever x-session-context is present in
// incoming metadata. It does not alter the handler path; publishing is fire-and-forget.
// Services opt in by adding this interceptor to their GetInterceptors() chain.
func SessionActivityInterceptor(serviceName string, producer SessionActivityProducer, topic string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return handler(ctx, req)
		}
		encoded := getMetadataValue(md, "x-session-context")
		if encoded == "" {
			return handler(ctx, req)
		}

		startTime := time.Now()
		resp, handlerErr := handler(ctx, req)
		duration := time.Since(startTime)

		if producer == nil {
			warnNilProducerOnce.Do(func() {
				log.Warn(ctx, "session_tracking_misconfigured",
					"x-session-context present but no Kafka producer configured; events will be dropped. "+
						"Register SessionInitializer and set orion.session_tracking.kafka_brokers.")
			})
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
			if err := producer.PublishAsync(topic, event); err != nil {
				log.Error(context.Background(), "session_activity_publish_failed",
					"error", err, "service", serviceName, "action", info.FullMethod)
			}
		}()
		return resp, handlerErr
	}
}

// warnNilProducerOnce emits a single warning when x-session-context arrives but
// no producer has been configured, telling the developer exactly what to fix.
var warnNilProducerOnce sync.Once

// GlobalSessionActivityProducer, GlobalSessionServiceName, and GlobalSessionActivityTopic
// are set by SessionInitializer.Init() before the gRPC server starts accepting connections.
// SessionInitializer.ReInit() is intentionally a no-op (same as HystrixInitializer), so
// these vars are written exactly once and are thereafter read-only — no synchronisation
// is required, matching the pattern used by every other Orion initializer.
var GlobalSessionActivityProducer SessionActivityProducer
var GlobalSessionServiceName string
var GlobalSessionActivityTopic = DefaultSessionActivityTopic

func SetGlobalSessionActivityProducer(p SessionActivityProducer) { GlobalSessionActivityProducer = p }
func SetGlobalSessionServiceName(name string)                     { GlobalSessionServiceName = name }
func SetGlobalSessionActivityTopic(topic string)                  { GlobalSessionActivityTopic = topic }

// GlobalSessionActivityInterceptor is a convenience interceptor that delegates to
// SessionActivityInterceptor using the globals set by SessionInitializer.
// Add this to GetInterceptors() in any Orion service that requires session tracking.
func GlobalSessionActivityInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		name := GlobalSessionServiceName
		if name == "" {
			name = "unknown-service"
		}
		return SessionActivityInterceptor(name, GlobalSessionActivityProducer, GlobalSessionActivityTopic)(ctx, req, info, handler)
	}
}

func getMetadataValue(md metadata.MD, key string) string {
	if values := md.Get(key); len(values) > 0 {
		return values[0]
	}
	return ""
}
