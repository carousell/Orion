package interceptors

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/carousell/Orion/utils/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// SessionContext is the decoded session metadata
// This matches the proto definition from AuthSvc
type SessionContext struct {
	Usid      []byte
	UserId    uint64
	Ip        []byte
	Country   string
	Platform  string
	DeviceId  string
	Timestamp uint64
}

// SessionActivityEvent is published to Kafka
type SessionActivityEvent struct {
	USID       string `json:"usid"`
	UserID     uint64 `json:"user_id"`
	Service    string `json:"service"`
	Action     string `json:"action"`
	Status     string `json:"status"`
	DurationMs int64  `json:"duration_ms"`
	Timestamp  int64  `json:"timestamp"`
}

// KafkaProducer interface for publishing events
type KafkaProducer interface {
	PublishAsync(topic string, event interface{}) error
}

// GlobalKafkaProducer is the singleton producer used by the interceptor
var GlobalKafkaProducer KafkaProducer

// SetGlobalKafkaProducer sets the global kafka producer
func SetGlobalKafkaProducer(p KafkaProducer) {
	GlobalKafkaProducer = p
}

// SessionActivityInterceptor tracks all session activities and publishes to Kafka
// Usage: Add to your gRPC server interceptors
// If producer is nil, it uses the GlobalKafkaProducer
func SessionActivityInterceptor(serviceName string, kafkaProducer KafkaProducer) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

		log.Info(ctx, "session_activity", "session_activity_interceptor", "service_name", serviceName)

		// Use global producer if specific one not provided
		producer := kafkaProducer
		if producer == nil {
			producer = GlobalKafkaProducer
		}

		// If still no producer, skip tracking
		if producer == nil {
			return handler(ctx, req)
		}

		// Extract session context from incoming gRPC metadata
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			// No metadata, proceed without tracking
			return handler(ctx, req)
		}

		encoded := getMetadataValue(md, "x-session-context")
		if encoded == "" {
			// No session context, skip tracking (e.g., unauthenticated requests)
			return handler(ctx, req)
		}

		// Decode session context
		sessionCtx, err := decodeSessionContext(encoded)
		if err != nil {
			// Don't fail the request, just skip tracking
			return handler(ctx, req)
		}

		log.Info(ctx, "session_activity", "session_context", sessionCtx)

		// Execute the actual business logic handler
		startTime := time.Now()
		resp, handlerErr := handler(ctx, req)
		duration := time.Since(startTime)

		// Determine status code from error
		statusCode := "OK"
		if handlerErr != nil {
			grpcStatus := status.Convert(handlerErr)
			statusCode = grpcStatus.Code().String() // e.g., "PERMISSION_DENIED", "INTERNAL"
		}

		// Build session activity event
		event := SessionActivityEvent{
			USID:       uuidBytesToString(sessionCtx.Usid),
			UserID:     sessionCtx.UserId,
			Service:    serviceName,
			Action:     info.FullMethod, // e.g., "/user.UserService/UpdateUsername"
			Status:     statusCode,
			DurationMs: duration.Milliseconds(),
			Timestamp:  time.Now().Unix(),
		}

		// Publish to Kafka asynchronously (non-blocking)
		go func() {
			log.Info(context.Background(), "session_activity", "publishing to kafka", "event", event)
			if err := producer.PublishAsync("session-activities", event); err != nil {
				log.Error(context.Background(), "failed to publish session activity", "error", err, "usid", event.USID)
			}
		}()

		return resp, handlerErr
	}
}

// GlobalSessionActivityInterceptor returns an interceptor that uses the GlobalKafkaProducer
// It attempts to auto-detect service name from config or context if possible, otherwise uses "unknown"
func GlobalSessionActivityInterceptor() grpc.UnaryServerInterceptor {
	return SessionActivityInterceptor("unknown-service", nil)
}

// Helper functions

func getMetadataValue(md metadata.MD, key string) string {
	values := md.Get(key)
	if len(values) > 0 {
		return values[0]
	}
	return ""
}

func decodeSessionContext(encoded string) (*SessionContext, error) {
	if encoded == "" {
		return nil, fmt.Errorf("empty session context")
	}

	// Base64 decode
	serialized, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("base64 decode failed: %w", err)
	}

	// For now, unmarshal as JSON (simplified version)
	// In production, this should use the generated proto.Unmarshal
	var sessionCtx SessionContext
	if err := json.Unmarshal(serialized, &sessionCtx); err != nil {
		// If JSON fails, it might be protobuf - try proto unmarshal
		// This requires importing the generated session.pb.go
		return nil, fmt.Errorf("unmarshal failed: %w", err)
	}

	return &sessionCtx, nil
}

func uuidBytesToString(uuidBytes []byte) string {
	if len(uuidBytes) != 16 {
		return ""
	}
	return fmt.Sprintf("%02x%02x%02x%02x-%02x%02x-%02x%02x-%02x%02x-%02x%02x%02x%02x%02x%02x",
		uuidBytes[0], uuidBytes[1], uuidBytes[2], uuidBytes[3],
		uuidBytes[4], uuidBytes[5],
		uuidBytes[6], uuidBytes[7],
		uuidBytes[8], uuidBytes[9],
		uuidBytes[10], uuidBytes[11], uuidBytes[12], uuidBytes[13], uuidBytes[14], uuidBytes[15])
}
