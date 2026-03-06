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

// SessionActivityEvent is published to Kafka whenever x-session-track is present and
// x-session-context is non-empty. EncodedSessionContext is the raw base64-encoded value
// forwarded as-is; the consumer decodes it.
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

// SessionActivityInterceptor is a gRPC server interceptor that publishes a
// SessionActivityEvent to the given Kafka topic when:
//  1. x-session-track header is "true" (set by gateways for selected APIs), AND
//  2. x-session-context is present in incoming metadata.
//
// Strip-on-consume: After reading x-session-track and x-session-operation, the interceptor
// removes them from the incoming metadata BEFORE calling handler(). This is critical because
// ForwardMetadataInterceptor (a default client interceptor in DefaultClientInterceptors)
// copies ALL incoming metadata to outgoing metadata. By stripping these headers, we prevent
// session tracking from cascading into downstream service-to-service calls.
//
// The x-session-context header is intentionally NOT stripped — downstream services may need
// it for other purposes (e.g., per-call re-injection via cheaders.WithSessionTracking).
//
// Services opt in by adding this interceptor to their GetInterceptors() chain.
func SessionActivityInterceptor(serviceName string, producer SessionActivityProducer, topic string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return handler(ctx, req)
		}

		// Guard 1: only fire when a gateway explicitly opted this request in.
		// Internal service-to-service calls will NOT have this header because the
		// interceptor strips it before calling handler(), and ForwardMetadataInterceptor
		// only sees the stripped metadata.
		if getMetadataValue(md, "x-session-track") != "true" {
			return handler(ctx, req)
		}

		// Guard 2: encoded session context must be present.
		encoded := getMetadataValue(md, "x-session-context")
		if encoded == "" {
			log.Info(ctx, "session_activity_interceptor", "service", serviceName,
				"action", info.FullMethod, "msg", "x-session-track present but x-session-context empty, skipping")
			return handler(ctx, req)
		}

		// Read optional operation label before stripping.
		operation := getMetadataValue(md, "x-session-operation")
		if operation == "" {
			operation = info.FullMethod
		}

		// ── Strip consumed headers so ForwardMetadataInterceptor won't relay them ──
		// This is the same mutation mechanism used by ObservabilityGatewayServerInterceptor
		// in go-utils/cinterceptors, proven safe with ForwardMetadataInterceptor.
		strippedMD := md.Copy()
		strippedMD.Delete("x-session-track")
		strippedMD.Delete("x-session-operation")
		ctx = metadata.NewIncomingContext(ctx, strippedMD)

		// ── Run handler with stripped ctx ──
		startTime := time.Now()
		resp, handlerErr := handler(ctx, req)
		duration := time.Since(startTime)

		log.Info(ctx, "session_activity_interceptor", "service", serviceName,
			"action", operation, "duration_ms", duration.Milliseconds())

		if producer == nil {
			warnNilProducerOnce.Do(func() {
				log.Error(ctx, "session_tracking_misconfigured",
					"x-session-track present but no Kafka producer configured; events will be dropped. "+
						"Register SessionInitializer and set orion.session_tracking.kafka_brokers.")
			})
			return resp, handlerErr
		}

		statusCode := "OK"
		if handlerErr != nil {
			log.Error(ctx, "session_activity_interceptor", "service", serviceName,
				"action", operation, "error", handlerErr)
			statusCode = status.Convert(handlerErr).Code().String()
		}
		event := SessionActivityEvent{
			EncodedSessionContext: encoded,
			Service:               serviceName,
			Action:                operation,
			Status:                statusCode,
			DurationMs:            duration.Milliseconds(),
			Timestamp:             time.Now().Unix(),
		}
		log.Info(ctx, "session_activity_interceptor", "service", serviceName,
			"action", operation, "event", event)
		go func() {
			if err := producer.PublishAsync(topic, event); err != nil {
				log.Error(context.Background(), "session_activity_publish_failed",
					"error", err, "service", serviceName, "action", operation)
			}
		}()
		return resp, handlerErr
	}
}

// warnNilProducerOnce emits a single warning when x-session-track arrives but
// no producer has been configured, telling the developer exactly what to fix.
var warnNilProducerOnce sync.Once

// GlobalSessionActivityProducer, GlobalSessionServiceName, and GlobalSessionActivityTopic
// are set by SessionInitializer.Init() before the gRPC server starts accepting connections.
// SessionInitializer.ReInit() is intentionally a no-op (same as HystrixInitializer), so
// these vars are written exactly once and are thereafter read-only — no synchronisation
// is required, matching the pattern used by every other Orion initializer.
var (
	GlobalSessionActivityProducer SessionActivityProducer
	GlobalSessionServiceName      string
	GlobalSessionActivityTopic    = DefaultSessionActivityTopic
)

func SetGlobalSessionActivityProducer(p SessionActivityProducer) { GlobalSessionActivityProducer = p }
func SetGlobalSessionServiceName(name string)                    { GlobalSessionServiceName = name }
func SetGlobalSessionActivityTopic(topic string)                 { GlobalSessionActivityTopic = topic }

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
