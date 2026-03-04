package orion

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/carousell/Orion/interceptors"
	"github.com/carousell/Orion/utils/log"
	"github.com/carousell/go-utils/kafka"
)

// kafkaProducerInitTimeout caps how long Init waits for sarama.NewAsyncProducer.
// sarama.NewAsyncProducer calls NewClient which performs a real TCP connection + metadata
// fetch. With Sarama's defaults (DialTimeout=30s, Metadata.Retry.Max=3) an unreachable
// broker stalls for ~120 seconds. We run it in a goroutine and abandon after this
// deadline so service startup is never blocked by Kafka.
// The orphaned goroutine exits on its own when Sarama's own dial timeout fires.
const kafkaProducerInitTimeout = 10 * time.Second

// kafkaProduceTimeout caps how long PublishAsync waits to enqueue one message.
// kafka.Producer.Produce does a blocking channel send onto Sarama's input channel
// (default buffer: 256 messages). When Kafka is down the buffer fills and every
// Produce call blocks indefinitely — the context parameter is ignored by go-utils/kafka.
// Without this cap, each interceptor goroutine that calls PublishAsync would leak for the
// entire duration of the outage. With the cap the goroutine returns an error after
// kafkaProduceTimeout and the interceptor logs a single "dropped event" line.
// The inner goroutine that wraps Produce is still blocked on the Sarama channel, but it
// self-resolves when Kafka recovers and the channel drains.
const kafkaProduceTimeout = 5 * time.Second

// rawProducer is the subset of kafka.Producer used internally.
// Defined as an interface so tests can inject fakes without touching go-utils/kafka.
type rawProducer interface {
	Run()
	Produce(ctx context.Context, topic string, key []byte, msg []byte) error
}

// SessionInitializer returns an Initializer that wires up the Kafka producer used by
// GlobalSessionActivityInterceptor. Services that want session tracking register this
// initializer and add GlobalSessionActivityInterceptor to their interceptor chain.
//
// Configuration (via viper / config file):
//
//	orion.session_tracking.kafka_brokers: ["broker1:9092", "broker2:9092"]
//	orion.session_tracking.kafka_topic:   "session-activities"  # optional, has default
//
// If kafka_brokers is not set, Init is a no-op and the interceptor silently drops events.
// A Kafka connection failure is non-fatal: the service starts but session tracking
// is disabled and a warning is emitted on the first request that carries x-session-context.
func SessionInitializer() Initializer {
	return &sessionInitializer{}
}

// sessionInitializer has no exported state — identical to hystrixInitializer.
// All configuration is written to package-level globals in Init() before the
// gRPC server starts accepting connections, so no synchronisation is needed.
//
// newProducer and initTimeout are zero-valued in production and overridden in tests
// to avoid real Kafka connections and slow timer waits.
type sessionInitializer struct {
	newProducer func(brokers []string, opts ...kafka.Option) (rawProducer, error)
	initTimeout time.Duration
}

func (s *sessionInitializer) Init(svr Server) error {
	cfg := svr.GetOrionConfig().SessionTrackingConfig
	brokers := cfg.KafkaBrokers
	if len(brokers) == 0 {
		log.Debug(context.Background(), "session_tracking", "kafka brokers not configured, skipping")
		return nil
	}

	topic := cfg.KafkaTopic
	if topic == "" {
		topic = interceptors.DefaultSessionActivityTopic
	}

	serviceName := svr.GetOrionConfig().OrionServerName
	if serviceName == "" {
		serviceName = "unknown-service"
	}

	factory := s.newProducer
	if factory == nil {
		factory = func(brokers []string, opts ...kafka.Option) (rawProducer, error) {
			return kafka.NewProducer(brokers, opts...)
		}
	}

	initTimeout := s.initTimeout
	if initTimeout == 0 {
		initTimeout = kafkaProducerInitTimeout
	}

	// sarama.NewAsyncProducer (called inside kafka.NewProducer) dials every broker to
	// fetch cluster metadata. With default Sarama timeouts (DialTimeout=30s,
	// Metadata.Retry.Max=3) an unreachable cluster stalls for up to ~120 seconds.
	// Run the call in a goroutine and abandon it after initTimeout so the service
	// always starts promptly. The abandoned goroutine exits on its own once Sarama's
	// internal dial timeout fires.
	type producerResult struct {
		p   rawProducer
		err error
	}
	ch := make(chan producerResult, 1)
	go func() {
		p, err := factory(brokers,
			kafka.WithMaxRetries(3),
			kafka.WithErrorHandler(func(err error) {
				log.Error(context.Background(), "session_tracking", "kafka producer error", "error", err)
			}),
		)
		ch <- producerResult{p, err}
	}()

	var producer rawProducer
	select {
	case res := <-ch:
		if res.err != nil {
			// Non-fatal: the service starts without session tracking. The interceptor
			// emits a single misconfiguration warning on the first request that carries
			// x-session-context.
			log.Error(context.Background(), "session_tracking",
				"failed to create kafka producer; session tracking disabled", "error", res.err)
			return nil
		}
		producer = res.p
	case <-time.After(initTimeout):
		log.Error(context.Background(), "session_tracking",
			"kafka producer creation timed out; session tracking disabled",
			"timeout", initTimeout.String())
		return nil
	}
	producer.Run()

	interceptors.SetGlobalSessionActivityProducer(&sessionProducerAdapter{producer: producer, defaultTopic: topic})
	interceptors.SetGlobalSessionServiceName(serviceName)
	interceptors.SetGlobalSessionActivityTopic(topic)
	log.Info(context.Background(), "session_tracking", "initialized",
		"brokers", brokers, "topic", topic, "service", serviceName)
	return nil
}

// ReInit is intentionally a no-op.
// Kafka broker or topic changes require a redeployment (new process → fresh Init).
// This mirrors HystrixInitializer which also skips ReInit with the comment
// "do nothing, can't be reinited".
func (s *sessionInitializer) ReInit(svr Server) error {
	return nil
}

// sessionProducerAdapter adapts go-utils kafka.Producer to interceptors.SessionActivityProducer.
// produceTimeout is zero-valued in production (falls back to kafkaProduceTimeout) and
// overridden in tests to avoid slow timer waits.
type sessionProducerAdapter struct {
	producer       rawProducer
	defaultTopic   string
	produceTimeout time.Duration
}

func (a *sessionProducerAdapter) PublishAsync(topic string, event interface{}) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	t := topic
	if t == "" {
		t = a.defaultTopic
	}

	timeout := a.produceTimeout
	if timeout == 0 {
		timeout = kafkaProduceTimeout
	}

	// kafka.Producer.Produce sends onto Sarama's buffered input channel (default 256
	// slots). When Kafka is unavailable the channel fills and Produce blocks indefinitely
	// — go-utils/kafka ignores the context passed to Produce. Without a cap the
	// interceptor goroutine (go func() { PublishAsync(...) }()) leaks for the duration
	// of the outage, accumulating memory proportional to traffic × outage length.
	//
	// Wrapping in a goroutine + select bounds the caller to timeout.
	// The inner goroutine is still blocked on the Sarama channel but is bounded to
	// (traffic rate × timeout) goroutines and self-resolves when Kafka recovers.
	errCh := make(chan error, 1)
	go func() {
		errCh <- a.producer.Produce(context.Background(), t, nil, payload)
	}()
	select {
	case err := <-errCh:
		return err
	case <-time.After(timeout):
		return fmt.Errorf("session_tracking: kafka produce timed out after %s (kafka may be unavailable); dropping event", timeout)
	}
}
