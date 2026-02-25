package orion

import (
	"context"
	"encoding/json"

	"github.com/spf13/viper"

	"github.com/carousell/Orion/interceptors"
	"github.com/carousell/Orion/utils/log"
	"github.com/carousell/go-utils/kafka"
)

// SessionInitializerConfigKey* are viper keys for session tracking Kafka. If not set, Init is a no-op.
const (
	SessionInitializerConfigKeyBrokers = "orion.SessionTrackingKafkaBrokers"
	SessionInitializerConfigKeyTopic   = "orion.SessionTopic"
)

// SessionInitializer returns an Initializer that sets up the session activity Kafka producer
// using go-utils/kafka and globals used by GlobalSessionActivityInterceptor. Only services
// that opt in to session tracking (e.g. UserSvc) should add this initializer. If
// orion.SessionTrackingKafkaBrokers is not configured, Init returns nil without error (no-op).
func SessionInitializer() Initializer {
	return &sessionInitializer{}
}

type sessionInitializer struct{}

func (s *sessionInitializer) Init(svr Server) error {
	brokers := viper.GetStringSlice(SessionInitializerConfigKeyBrokers)
	if len(brokers) == 0 {
		log.Debug(context.Background(), "session_tracking", "kafka brokers not configured, skipping")
		return nil
	}
	topic := viper.GetString(SessionInitializerConfigKeyTopic)
	if topic == "" {
		topic = interceptors.DefaultSessionActivityTopic
	}
	serviceName := viper.GetString("orion.ServiceName")
	if serviceName == "" {
		serviceName = svr.GetOrionConfig().OrionServerName
	}
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

	wrapper := &sessionActivityProducerWrapper{producer: producer, defaultTopic: topic}
	interceptors.SetGlobalSessionActivityProducer(wrapper)
	interceptors.SetGlobalSessionServiceName(serviceName)
	log.Info(context.Background(), "session_tracking", "initialized", "brokers", brokers, "topic", topic, "service_name", serviceName)
	return nil
}

func (s *sessionInitializer) ReInit(svr Server) error {
	return s.Init(svr)
}

// sessionActivityProducerWrapper adapts go-utils kafka.Producer to interceptors.SessionActivityProducer.
type sessionActivityProducerWrapper struct {
	producer     *kafka.Producer
	defaultTopic string
}

func (w *sessionActivityProducerWrapper) PublishAsync(topic string, event interface{}) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	t := topic
	if t == "" {
		t = w.defaultTopic
	}
	return w.producer.Produce(context.Background(), t, nil, payload)
}
