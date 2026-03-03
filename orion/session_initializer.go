package orion

import (
	"context"
	"encoding/json"

	"github.com/carousell/Orion/interceptors"
	"github.com/carousell/Orion/utils/log"
	"github.com/carousell/go-utils/kafka"
)

// SessionInitializer returns an Initializer that wires up the Kafka producer used by
// GlobalSessionActivityInterceptor. Services that want session tracking register this
// initializer and add GlobalSessionActivityInterceptor to their interceptor chain.
// If orion.SessionTrackingKafkaBrokers is not set, Init is a no-op.
func SessionInitializer() Initializer {
	return &sessionInitializer{}
}

type sessionInitializer struct{}

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

	producer, err := kafka.NewProducer(brokers,
		kafka.WithMaxRetries(3),
		kafka.WithErrorHandler(func(err error) {
			log.Error(context.Background(), "session_tracking", "kafka producer error", "error", err)
		}),
	)
	if err != nil {
		log.Error(context.Background(), "session_tracking", "failed to create kafka producer", "error", err)
		return err
	}
	producer.Run()

	interceptors.SetGlobalSessionActivityProducer(&sessionProducerAdapter{producer: producer, defaultTopic: topic})
	interceptors.SetGlobalSessionServiceName(serviceName)
	interceptors.SetGlobalSessionActivityTopic(topic)
	log.Info(context.Background(), "session_tracking", "initialized", "brokers", brokers, "topic", topic, "service", serviceName)
	return nil
}

func (s *sessionInitializer) ReInit(svr Server) error {
	return s.Init(svr)
}

// sessionProducerAdapter adapts go-utils kafka.Producer to interceptors.SessionActivityProducer.
type sessionProducerAdapter struct {
	producer     *kafka.Producer
	defaultTopic string
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
	return a.producer.Produce(context.Background(), t, nil, payload)
}
