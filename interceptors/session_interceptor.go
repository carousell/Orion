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

// SessionActivityEvent is published to Kafka for each tracked session request.
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

// SessionActivityInterceptor publishes a SessionActivityEvent to Kafka when
// x-session-track="true" and x-session-context are present in incoming metadata.
// x-session-track and x-session-operation are stripped before calling the handler
// to prevent forwarding to downstream services.
func SessionActivityInterceptor(serviceName string, producer SessionActivityProducer, topic string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return handler(ctx, req)
		}

		// Only track requests opted in by a gateway.
		if getMetadataValue(md, "x-session-track") != "true" {
			return handler(ctx, req)
		}

		// Session context must be present.
		encoded := getMetadataValue(md, "x-session-context")
		if encoded == "" {
			log.Info(ctx, "session_activity_interceptor", "service", serviceName,
				"action", info.FullMethod, "msg", "x-session-track present but x-session-context empty, skipping")
			return handler(ctx, req)
		}

		// Use operation label if provided, otherwise fall back to method name.
		operation := getMetadataValue(md, "x-session-operation")
		if operation == "" {
			operation = info.FullMethod
		}

		// Strip tracking headers to prevent forwarding to downstream services.
		strippedMD := md.Copy()
		strippedMD.Delete("x-session-track")
		strippedMD.Delete("x-session-operation")
		ctx = metadata.NewIncomingContext(ctx, strippedMD)

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

var warnNilProducerOnce sync.Once

// Globals set by SessionInitializer.Init() before the server starts; read-only after.
var (
	GlobalSessionActivityProducer SessionActivityProducer
	GlobalSessionServiceName      string
	GlobalSessionActivityTopic    = DefaultSessionActivityTopic
)

func SetGlobalSessionActivityProducer(p SessionActivityProducer) { GlobalSessionActivityProducer = p }
func SetGlobalSessionServiceName(name string)                    { GlobalSessionServiceName = name }
func SetGlobalSessionActivityTopic(topic string)                 { GlobalSessionActivityTopic = topic }

// GlobalSessionActivityInterceptor delegates to SessionActivityInterceptor using
// the globals set by SessionInitializer.
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
