package kafka

import (
	"context"
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/carousell/Orion/utils/errors"
	"github.com/carousell/Orion/utils/log"
	"github.com/carousell/Orion/utils/spanutils"
	"github.com/carousell/go-utils/utils/hystrixwrapper"
	"github.com/golang/protobuf/proto"
)

// Producer is a wrapper interface around sarama SyncProducer
type Producer interface {
	Publish(ctx context.Context, topic string, key string, msg proto.Message) error
	Close() error
}

type producer struct {
	syncProducer sarama.SyncProducer
}

// NewProducer creates and returns a new Producer given config
func NewProducer(config ProducerConfig) (Producer, error) {
	defaultConfig := sarama.NewConfig()
	defaultConfig.ClientID = config.ClientID
	defaultConfig.Version = config.KafkaVersion
	defaultConfig.Producer.RequiredAcks = config.RequiredAcks
	defaultConfig.Producer.Return.Errors = true
	defaultConfig.Producer.Return.Successes = true
	defaultConfig.Net.DialTimeout = config.KafkaSocketTimeout
	defaultConfig.Net.ReadTimeout = config.KafkaSocketTimeout
	defaultConfig.Net.WriteTimeout = config.KafkaSocketTimeout
	syncProducer, err := sarama.NewSyncProducer(config.Brokers, defaultConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new kafka sync producer from client")
	}
	producer := &producer{
		syncProducer: syncProducer,
	}
	return producer, nil
}

// Publish publishes a proto message to the provided topic
func (p *producer) Publish(ctx context.Context, topic string, key string, msg proto.Message) error {
	name := "KafkaPublish"
	span, ctx := spanutils.NewDatastoreSpan(ctx, name, "Kafka")
	defer span.Finish()

	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to marshal proto message %v", msg))
	}

	saramaMsg := sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(msgBytes),
	}

	var partition int32
	var offset int64
	var e error
	err = hystrixwrapper.DoCWithOptions(ctx, name, func(ctx context.Context) error {
		partition, offset, e = p.syncProducer.SendMessage(&saramaMsg)
		return errors.Wrap(e, fmt.Sprintf("failed to send message %v", saramaMsg))
	}, nil, hystrixwrapper.WithRecover())
	if err != nil {
		span.SetError(err.Error())
		return err
	}
	log.Info(ctx, "kafka_action", "Publish", "topic", topic, "partition", partition, "offset", offset)
	return nil
}

// Close closes sarama SyncProducer
func (p *producer) Close() error {
	if p.syncProducer == nil {
		return nil
	}
	return errors.Wrap(p.syncProducer.Close(), "failed to close kafka sync producer")
}
