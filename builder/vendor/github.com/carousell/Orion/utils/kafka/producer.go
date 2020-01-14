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
	config        *Config
	channel       chan *sarama.ProducerMessage
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
		config:        &c,
		channel:       make(chan *sarama.ProducerMessage, c.QueueBufferSize),
	}, nil
}

// Run tells the producer to start accepting messages to publish to Kafka
func (p *Producer) Run() {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				notifier.NotifyWithLevel(errors.Wrap(fmt.Errorf("%v", r), "panic in Kafka producer"), "critical")
			}
		}()

		for {
			select {
			case msg, ok := <-p.channel:
				if !ok {
					// producer has been stopped at user level
					return
				}
				p.asyncProducer.Input() <- msg
			case err, ok := <-p.asyncProducer.Errors():
				if !ok {
					// producer has been stopped at Sarama level
					return
				}
				notifier.Notify(errors.Wrap(err, "failed to produce Kafka message"))
			}
		}
	}()
}

// Produce sends a message to a particular Kafka topic
func (p *Producer) Produce(ctx context.Context, topic string, key string, msg []byte) error {
	if p.channel == nil {
		return errors.New("message discarded because producer has not been started")
	}

	p.channel <- &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(msg),
	}
	return nil
}

// Close stops the producer from accepting and sending any new messages
func (p *Producer) Close() error {
	if p.channel != nil {
		close(p.channel)
	}
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
