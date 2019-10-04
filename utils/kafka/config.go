package kafka

import (
	"time"

	"github.com/Shopify/sarama"
)

type ProducerConfig struct {
	Brokers            []string
	KafkaSocketTimeout time.Duration
	ClientID           string
	KafkaVersion       sarama.KafkaVersion
	RequiredAcks       sarama.RequiredAcks
}

type ConsumerGroupConfig struct {
	Brokers            []string
	KafkaSocketTimeout time.Duration
	ClientID           string
	KafkaVersion       sarama.KafkaVersion
	Name               string
	Topic              string
	CommitInterval     time.Duration
}
