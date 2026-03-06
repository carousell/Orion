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

// SessionActivityEvent is the payload published to Kafka for each tracked session request.
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
// SessionActivityEvent to Kafka when both x-session-track="true" and x-session-context
// are present in the incoming metadata.
//
// x-session-track and x-session-operation are stripped from metadata before calling the
// handler, so they won't be forwarded to downstream services. x-session-context is kept
// so downstream services can still use it if needed.
//
// Add this to GetInterceptors() in any service that requires session tracking.
func SessionActivityInterceptor(serviceName string, producer SessionActivityProducer, topic string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return handler(ctx, req)
		}

		// Only track requests explicitly opted in by a gateway.
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

		// Strip session tracking headers so they aren't forwarded to downstream services.
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

// warnNilProducerOnce ensures the misconfiguration warning is logged only once.
var warnNilProducerOnce sync.Once

// These globals are set once by SessionInitializer.Init() before the server starts.
// They are read-only after that, so no synchronisation is needed.
var (
	GlobalSessionActivityProducer SessionActivityProducer
	GlobalSessionServiceName      string
	GlobalSessionActivityTopic    = DefaultSessionActivityTopic
)

func SetGlobalSessionActivityProducer(p SessionActivityProducer) { GlobalSessionActivityProducer = p }
func SetGlobalSessionServiceName(name string)                    { GlobalSessionServiceName = name }
func SetGlobalSessionActivityTopic(topic string)                 { GlobalSessionActivityTopic = topic }

// GlobalSessionActivityInterceptor wraps SessionActivityInterceptor using the globals
// set by SessionInitializer. Add this to GetInterceptors() to enable session tracking.
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
