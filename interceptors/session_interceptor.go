package interceptors

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/carousell/Orion/utils/log"
)

// SessionContextKey is the context key for storing the decoded session context
type sessionContextKey struct{}

var SessionContextKey = sessionContextKey{}

// FromContext retrieves the SessionContext from the context
func FromContext(ctx context.Context) *SessionContext {
	val := ctx.Value(SessionContextKey)
	if val == nil {
		return nil
	}
	if sc, ok := val.(*SessionContext); ok {
		return sc
	}
	return nil
}

// SessionContext is the decoded session metadata
type SessionContext struct {
	Usid      []byte `json:"usid"`
	UserId    uint64 `json:"user_id"`
	Ip        []byte `json:"ip"`
	Country   string `json:"country"`
	Platform  string `json:"platform"`
	DeviceId  string `json:"device_id"`
	Timestamp uint64 `json:"timestamp"`
}

// Unmarshal decodes protobuf binary data into the SessionContext struct
// This reverse-engineers the AuthSvc proto definitions without adding dependencies.
func (s *SessionContext) Unmarshal(data []byte) error {
	for len(data) > 0 {
		tagWire, n := decodeVarint(data)
		if n <= 0 {
			break
		}
		tag := tagWire >> 3
		wire := tagWire & 0x07
		data = data[n:]

		switch wire {
		case 0: // Varint
			val, vn := decodeVarint(data)
			if vn <= 0 {
				return fmt.Errorf("malformed varint")
			}
			data = data[vn:]
			if tag == 2 {
				s.UserId = val
			} else if tag == 7 {
				s.Timestamp = val
			}
		case 2: // Length-delimited
			length, ln := decodeVarint(data)
			if ln <= 0 || int(length) > len(data) {
				return fmt.Errorf("malformed length")
			}
			data = data[ln:]
			chunk := data[:length]
			data = data[length:]
			switch tag {
			case 1:
				s.Usid = chunk
			case 3:
				s.Ip = chunk
			case 4:
				s.Country = string(chunk)
			case 5:
				s.Platform = string(chunk)
			case 6:
				s.DeviceId = string(chunk)
			}
		default:
			// Skip unknown wire types
			return fmt.Errorf("unsupported wire type %d", wire)
		}
	}
	return nil
}

// Marshal encodes SessionContext into protobuf binary data
func (s *SessionContext) Marshal() ([]byte, error) {
	var data []byte

	// Tag 1: Usid (bytes)
	if len(s.Usid) > 0 {
		data = append(data, encodeTag(1, 2)...)
		data = append(data, encodeVarint(uint64(len(s.Usid)))...)
		data = append(data, s.Usid...)
	}

	// Tag 2: UserId (varint)
	if s.UserId != 0 {
		data = append(data, encodeTag(2, 0)...)
		data = append(data, encodeVarint(s.UserId)...)
	}

	// Tag 3: Ip (bytes)
	if len(s.Ip) > 0 {
		data = append(data, encodeTag(3, 2)...)
		data = append(data, encodeVarint(uint64(len(s.Ip)))...)
		data = append(data, s.Ip...)
	}

	// Tag 4: Country (string)
	if len(s.Country) > 0 {
		data = append(data, encodeTag(4, 2)...)
		data = append(data, encodeVarint(uint64(len(s.Country)))...)
		data = append(data, []byte(s.Country)...)
	}

	// Tag 5: Platform (string)
	if len(s.Platform) > 0 {
		data = append(data, encodeTag(5, 2)...)
		data = append(data, encodeVarint(uint64(len(s.Platform)))...)
		data = append(data, []byte(s.Platform)...)
	}

	// Tag 6: DeviceId (string)
	if len(s.DeviceId) > 0 {
		data = append(data, encodeTag(6, 2)...)
		data = append(data, encodeVarint(uint64(len(s.DeviceId)))...)
		data = append(data, []byte(s.DeviceId)...)
	}

	// Tag 7: Timestamp (varint)
	if s.Timestamp != 0 {
		data = append(data, encodeTag(7, 0)...)
		data = append(data, encodeVarint(s.Timestamp)...)
	}

	return data, nil
}

// SessionActivityEvent is published to Kafka
type SessionActivityEvent struct {
	USID       string                  `json:"usid"`
	UserID     uint64                  `json:"user_id"`
	Service    string                  `json:"service"`
	Action     string                  `json:"action"`
	Status     string                  `json:"status"`
	DurationMs int64                   `json:"duration_ms"`
	Metadata   SessionActivityMetadata `json:"metadata"`
	Timestamp  int64                   `json:"timestamp"`
}

// SessionActivityMetadata contains contextual information for each activity
type SessionActivityMetadata struct {
	ClientIP string `json:"client_ip"`
	Country  string `json:"country"`
	Platform string `json:"platform"`
	DeviceID string `json:"device_id"`
}

// KafkaProducer interface for publishing events
type KafkaProducer interface {
	PublishAsync(topic string, event interface{}) error
}

// GlobalKafkaProducer is the singleton producer used by the interceptor
var GlobalKafkaProducer KafkaProducer

// GlobalServiceName is the singleton service name used by the interceptor
var GlobalServiceName string

// SetGlobalKafkaProducer sets the global kafka producer
func SetGlobalKafkaProducer(p KafkaProducer) {
	GlobalKafkaProducer = p
}

// SetGlobalServiceName sets the global service name
func SetGlobalServiceName(name string) {
	GlobalServiceName = name
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
			log.Error(ctx, "session_activity", "no_kafka_producer", "service_name", serviceName)
			return handler(ctx, req)
		}

		// Extract session context from incoming gRPC metadata
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			// No metadata, proceed without tracking
			log.Error(ctx, "session_activity", "no_metadata", "service_name", serviceName)
			return handler(ctx, req)
		}

		encoded := getMetadataValue(md, "x-session-context")
		if encoded == "" {
			// No session context, skip tracking (e.g., unauthenticated requests)
			log.Error(ctx, "session_activity", "no_session_context", "service_name", serviceName)
			return handler(ctx, req)
		}

		// Decode session context
		sessionCtx, err := decodeSessionContext(encoded)
		if err != nil {
			log.Error(ctx, "session_activity", "failed_to_decode_session_context", "service_name", serviceName, "error", err)
			// Don't fail the request, just skip tracking
			return handler(ctx, req)
		}

		// Store in context for downstream access
		ctx = context.WithValue(ctx, SessionContextKey, sessionCtx)

		// Check if we should log this activity

		// Check if we should log this activity
		// Reliance on "x-session-tracking" header from Gateway or upstream
		trackingHeader := getMetadataValue(md, "x-session-tracking")
		shouldLog := trackingHeader == "true"

		if !shouldLog {
			// Skip logging, but we still propagated the context
			return handler(ctx, req)
		}

		// Ensure propagation: Add header to outgoing context for downstream services
		ctx = metadata.AppendToOutgoingContext(ctx, "x-session-tracking", "true")

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

		// Build session activity event with metadata (PRD compliant)
		event := SessionActivityEvent{
			USID:       uuidBytesToString(sessionCtx.Usid),
			UserID:     sessionCtx.UserId,
			Service:    serviceName,
			Action:     info.FullMethod, // e.g., "/user.UserService/UpdateUsername"
			Status:     statusCode,
			DurationMs: duration.Milliseconds(),
			Metadata: SessionActivityMetadata{
				ClientIP: ipBytesToString(sessionCtx.Ip),
				Country:  sessionCtx.Country,
				Platform: sessionCtx.Platform,
				DeviceID: sessionCtx.DeviceId,
			},
			Timestamp: time.Now().Unix(),
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
// It uses the GlobalServiceName set by the SessionInitializer
func GlobalSessionActivityInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Use the global service name set by the initializer
		serviceName := GlobalServiceName
		if serviceName == "" {
			serviceName = "unknown-service"
		}

		// Delegate to the main SessionActivityInterceptor
		interceptor := SessionActivityInterceptor(serviceName, nil)
		return interceptor(ctx, req, info, handler)
	}
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
		return nil, fmt.Errorf("empty")
	}

	serialized, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	ctx := &SessionContext{}
	if err := ctx.Unmarshal(serialized); err != nil {
		// Fallback for legacy JSON support if needed
		if len(serialized) > 0 && serialized[0] == '{' {
			_ = json.Unmarshal(serialized, ctx)
		}
		return ctx, nil
	}
	return ctx, nil
}

func decodeVarint(data []byte) (uint64, int) {
	var val uint64
	var shift uint
	for i, b := range data {
		val |= uint64(b&0x7F) << shift
		if b&0x80 == 0 {
			return val, i + 1
		}
		shift += 7
	}
	return 0, 0
}

func encodeVarint(x uint64) []byte {
	var buf []byte
	for x >= 1<<7 {
		buf = append(buf, uint8(x&0x7f|0x80))
		x >>= 7
	}
	buf = append(buf, uint8(x))
	return buf
}

func encodeTag(fieldNumber int, wireType int) []byte {
	return encodeVarint(uint64(fieldNumber<<3 | wireType))
}

func uuidBytesToString(uuidBytes []byte) string {
	if len(uuidBytes) != 16 {
		return ""
	}
	// Same format as AuthSvc UUIDToString function
	return fmt.Sprintf("%02x%02x%02x%02x-%02x%02x-%02x%02x-%02x%02x-%02x%02x%02x%02x%02x%02x",
		uuidBytes[0], uuidBytes[1], uuidBytes[2], uuidBytes[3],
		uuidBytes[4], uuidBytes[5],
		uuidBytes[6], uuidBytes[7],
		uuidBytes[8], uuidBytes[9],
		uuidBytes[10], uuidBytes[11], uuidBytes[12], uuidBytes[13], uuidBytes[14], uuidBytes[15])
}

// ipBytesToString converts IP bytes back to string (mirrors AuthSvc IPToString)
func ipBytesToString(ipBytes []byte) string {
	if len(ipBytes) == 0 {
		return ""
	}
	// Same logic as AuthSvc IPToString function - handle both IPv4 and IPv6
	if len(ipBytes) == 4 {
		// IPv4
		return fmt.Sprintf("%d.%d.%d.%d", ipBytes[0], ipBytes[1], ipBytes[2], ipBytes[3])
	} else if len(ipBytes) == 16 {
		// IPv6 - use Go's net.IP for proper formatting
		return fmt.Sprintf("%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x",
			ipBytes[0], ipBytes[1], ipBytes[2], ipBytes[3],
			ipBytes[4], ipBytes[5], ipBytes[6], ipBytes[7],
			ipBytes[8], ipBytes[9], ipBytes[10], ipBytes[11],
			ipBytes[12], ipBytes[13], ipBytes[14], ipBytes[15])
	}
	return ""
}
