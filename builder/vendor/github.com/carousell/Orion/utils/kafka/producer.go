package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/Shopify/sarama"
	"github.com/carousell/Orion/utils/errors"
	"github.com/carousell/Orion/utils/errors/notifier"
)

// Producer is a Kafka producer based on Sarama
type Producer struct {
	asyncProducer sarama.AsyncProducer
	open          bool
}

// NewProducer creates a Kafka producer
func NewProducer(c Config) (*Producer, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.Flush.Frequency = c.FlushInterval
	saramaConfig.Producer.Retry.Max = c.MaxRetries

	if len(c.Brokers) == 0 {
		return nil, errors.New("must provide at least one Kafka broker")
	}

	asyncProducer, err := sarama.NewAsyncProducer(c.Brokers, saramaConfig)
	if err != nil {
		return nil, errors.Wrap(err, "error creating async producer")
	}

	return &Producer{
		asyncProducer: asyncProducer,
	}, nil
}

// Run tells the producer to start accepting messages to publish to Kafka
func (p *Producer) Run() {
	p.open = true
	go func() {
		defer func() {
			if r := recover(); r != nil {
				notifier.NotifyWithLevel(errors.Wrap(fmt.Errorf("%v", r), "panic in Kafka producer error handler"), "critical")
			}
		}()
		for {
			select {
			case err, ok := <-p.asyncProducer.Errors():
				if !ok {
					return
				}
				notifier.Notify(errors.Wrap(err, "failed to produce Kafka message"))
			}
		}
	}()
}

// Produce sends a message to a particular Kafka topic
func (p *Producer) Produce(ctx context.Context, topic string, key string, msg []byte) error {
	if !p.open || p.asyncProducer != nil {
		return errors.New("producer is closed")
	}
	saramaMsg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(msg),
	}
	p.asyncProducer.Input() <- saramaMsg
	return nil
}

// Close stops the producer from accepting and sending any new messages
func (p *Producer) Close() error {
	p.open = false
	err := p.asyncProducer.Close()
	if err != nil {
		return errors.Wrap(err, "failed to close Kafka async producer")
	}
	return nil
}

// Config contains Kafka connection parameters
type Config struct {
	Brokers         []string
	FlushInterval   time.Duration
	MaxRetries      int
	QueueBufferSize int
}
