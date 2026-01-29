package orion

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Shopify/sarama"
	"github.com/spf13/viper"

	"github.com/carousell/Orion/interceptors"
	"github.com/carousell/Orion/utils/log"
)

// SessionInitializer returns a Initializer implementation for Session Tracking
func SessionInitializer() Initializer {
	return &sessionInitializer{}
}

type sessionInitializer struct {
}

func (s *sessionInitializer) Init(svr Server) error {
	// Read config from Viper (Orion configuration)
	brokers := viper.GetStringSlice("orion.KafkaBrokers")
	topic := viper.GetString("orion.SessionTopic")

	if len(brokers) == 0 {
		log.Warn(context.Background(), "session_tracking", "kafka brokers not configured, skipping initialization")
		return nil
	}

	if topic == "" {
		topic = "session-activities" // Default topic
	}

	// Get service name from server config (same pattern as NewRelic initializer)
	serviceName := viper.GetString("orion.ServiceName")
	if serviceName == "" {
		serviceName = svr.GetOrionConfig().OrionServerName
	}
	if serviceName == "" {
		serviceName = "unknown-service"
	}

	config := sarama.NewConfig()
	config.Producer.Return.Errors = true // We want to log errors
	config.Producer.RequiredAcks = sarama.WaitForLocal
	config.Producer.Compression = sarama.CompressionSnappy
	config.Producer.Retry.Max = 3
	config.Producer.Timeout = 5 * time.Second

	producer, err := sarama.NewAsyncProducer(brokers, config)
	if err != nil {
		log.Error(context.Background(), "session_tracking", "failed to initialize kafka producer", "error", err)
		return err
	}

	// Start error logging goroutine
	go func() {
		for err := range producer.Errors() {
			log.Error(context.Background(), "session_tracking", "kafka producer error", "error", err.Err, "msg", err.Msg)
		}
	}()

	// Set the global producer in interceptors package
	interceptors.SetGlobalKafkaProducer(&orionKafkaWrapper{
		producer: producer,
		topic:    topic,
	})

	// Set the global service name in interceptors package
	interceptors.SetGlobalServiceName(serviceName)

	log.Info(context.Background(), "session_tracking", "initialized", "brokers", brokers, "topic", topic, "service_name", serviceName)
	return nil
}

func (s *sessionInitializer) ReInit(svr Server) error {
	// For now, re-init is same as init (might create new producer)
	return s.Init(svr)
}

// orionKafkaWrapper wraps sarama producer to match interceptors.KafkaProducer interface
type orionKafkaWrapper struct {
	producer sarama.AsyncProducer
	topic    string
}

func (w *orionKafkaWrapper) PublishAsync(topic string, event interface{}) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	targetTopic := topic
	if targetTopic == "" {
		targetTopic = w.topic
	}

	message := &sarama.ProducerMessage{
		Topic: targetTopic,
		Value: sarama.ByteEncoder(payload),
	}

	log.Info(context.Background(), "session_tracking", "publishing event", "topic", targetTopic, "event", string(payload))

	w.producer.Input() <- message
	return nil
}
