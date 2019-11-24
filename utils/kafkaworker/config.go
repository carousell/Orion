package kafkaworker

import (
	"time"

	"github.com/Shopify/sarama"
	"github.com/carousell/Orion/utils/kafka"
)

// Config holds options for a Kafka based worker
type Config struct {
	Brokers            []string
	KafkaSocketTimeout time.Duration
	ClientID           string
	KafkaVersion       sarama.KafkaVersion
	CGName             string
	Topic              string
	CommitInterval     time.Duration
	// Infinite retries can be specified with Retries = -1
	Retries       int
	RetryInterval time.Duration
	// If a worker specifies sideline options, failed messages
	// will be sent to the sideline topic
	Sideline *SidelineConfig
	Enabled  bool
}

// SidelineConfig holds options for sending messages to a sideline topic
type SidelineConfig struct {
	SidelineTopic string
	RequiredAcks  sarama.RequiredAcks
}

func newKafkaCGConfig(config Config) kafka.ConsumerGroupConfig {
	return kafka.ConsumerGroupConfig{
		Brokers:            config.Brokers,
		KafkaSocketTimeout: config.KafkaSocketTimeout,
		ClientID:           config.ClientID,
		KafkaVersion:       config.KafkaVersion,
		Name:               config.CGName,
		Topic:              config.Topic,
		CommitInterval:     config.CommitInterval,
	}
}

func newKafkaProducerConfig(config Config) kafka.ProducerConfig {
	if config.Sideline == nil {
		return kafka.ProducerConfig{}
	}
	return kafka.ProducerConfig{
		Brokers:            config.Brokers,
		KafkaSocketTimeout: config.KafkaSocketTimeout,
		ClientID:           config.ClientID,
		KafkaVersion:       config.KafkaVersion,
		RequiredAcks:       config.Sideline.RequiredAcks,
	}
}
