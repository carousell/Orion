package orion

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Shopify/sarama"
	"github.com/carousell/Orion/interceptors"
	"github.com/carousell/Orion/utils/log"
)

// kafkaProducerInitTimeout caps how long Init waits for the Kafka producer to connect.
const kafkaProducerInitTimeout = 10 * time.Second

// kafkaProduceTimeout caps how long PublishAsync waits to enqueue a message.
const kafkaProduceTimeout = 5 * time.Second

// rawProducer abstracts the Kafka producer for testability.
type rawProducer interface {
	Run()
	Produce(ctx context.Context, topic string, key []byte, msg []byte) error
}

// saramaProducer wraps sarama.AsyncProducer to implement rawProducer.
type saramaProducer struct {
	p            sarama.AsyncProducer
	errorHandler func(error)
}

func newSaramaProducer(brokers []string, errorHandler func(error)) (rawProducer, error) {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Errors = true
	cfg.Producer.Return.Successes = false
	cfg.Producer.Retry.Max = 3
	p, err := sarama.NewAsyncProducer(brokers, cfg)
	if err != nil {
		return nil, err
	}
	return &saramaProducer{p: p, errorHandler: errorHandler}, nil
}

func (sp *saramaProducer) Run() {
	go func() {
		for err := range sp.p.Errors() {
			if sp.errorHandler != nil {
				sp.errorHandler(err)
			}
		}
	}()
}

func (sp *saramaProducer) Produce(_ context.Context, topic string, key []byte, msg []byte) error {
	var k sarama.Encoder
	if len(key) > 0 {
		k = sarama.ByteEncoder(key)
	}
	sp.p.Input() <- &sarama.ProducerMessage{
		Topic: topic,
		Key:   k,
		Value: sarama.ByteEncoder(msg),
	}
	return nil
}

// SessionInitializer wires up the Kafka producer for session tracking.
// Requires orion.SessionTracking.KafkaBrokers to be configured; otherwise Init is a no-op.
func SessionInitializer() Initializer {
	return &sessionInitializer{}
}

// sessionInitializer fields are zero-valued in production and overridden in tests.
type sessionInitializer struct {
	newProducer func(brokers []string, errorHandler func(error)) (rawProducer, error)
	initTimeout time.Duration
}

func (s *sessionInitializer) Init(svr Server) error {
	cfg := svr.GetOrionConfig().SessionTrackingConfig
	brokers := cfg.KafkaBrokers
	log.Info(context.Background(), "session_tracking", "kafka brokers", "brokers", brokers)
	if len(brokers) == 0 {
		log.Error(context.Background(), "session_tracking", "kafka brokers not configured, skipping")
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
		factory = newSaramaProducer
	}

	initTimeout := s.initTimeout
	if initTimeout == 0 {
		initTimeout = kafkaProducerInitTimeout
	}

	// Create producer in a goroutine to avoid blocking startup if Kafka is slow.
	type producerResult struct {
		p   rawProducer
		err error
	}
	ch := make(chan producerResult, 1)
	go func() {
		p, err := factory(brokers, func(err error) {
			log.Error(context.Background(), "session_tracking", "kafka producer error", "error", err)
		})
		ch <- producerResult{p, err}
	}()

	var producer rawProducer
	select {
	case res := <-ch:
		if res.err != nil {
			// Non-fatal: session tracking disabled.
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
	log.Info(context.Background(), "session_tracking", "kafka producer run")

	interceptors.SetGlobalSessionActivityProducer(&sessionProducerAdapter{producer: producer, defaultTopic: topic})
	interceptors.SetGlobalSessionServiceName(serviceName)
	interceptors.SetGlobalSessionActivityTopic(topic)
	log.Info(context.Background(), "session_tracking", "initialized",
		"brokers", brokers, "topic", topic, "service", serviceName)
	return nil
}

// ReInit is a no-op; Kafka config changes require a redeployment.
func (s *sessionInitializer) ReInit(svr Server) error {
	log.Info(context.Background(), "session_tracking", "reinit")
	return nil
}

// sessionProducerAdapter adapts rawProducer to SessionActivityProducer.
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

	log.Info(context.Background(), "session_tracking", "payload", string(payload))

	// Produce with timeout to avoid blocking if Kafka is down.
	errCh := make(chan error, 1)
	go func() {
		errCh <- a.producer.Produce(context.Background(), t, nil, payload)
	}()
	select {
	case err := <-errCh:
		log.Info(context.Background(), "session_tracking", "kafka produce", "error", err)
		return err
	case <-time.After(timeout):
		log.Info(context.Background(), "session_tracking", "kafka produce timed out", "timeout", timeout)
		return fmt.Errorf("session_tracking: kafka produce timed out after %s (kafka may be unavailable); dropping event", timeout)
	}
}
