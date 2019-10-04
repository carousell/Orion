package kafkaworker

import (
	"context"
	"fmt"
	"time"

	"github.com/Shopify/sarama"
	"github.com/carousell/Orion/utils/errors"
	"github.com/carousell/Orion/utils/errors/notifier"
	"github.com/carousell/Orion/utils/kafka"
	"github.com/golang/protobuf/proto"
)

// TaskHandler processes tasks picked up from the queue
type TaskHandler interface {
	// Handle processes a message from the worker
	Handle(ctx context.Context, msg proto.Message) error
	// PayloadProto returns an example message object. It is used to deserialize
	// the raw bytes returned by the worker into a proto message that can
	// be used by the handler method
	PayloadProto() proto.Message
}

// A Worker processes messages for a particular topic, and also provides
// a method to schedule new messages
type Worker interface {
	Consume()
	Close() error
}

type worker struct {
	config           Config
	consumerGroup    kafka.ConsumerGroup
	sidelineProducer kafka.Producer

	taskHandler TaskHandler
}

func (w *worker) Consume() {
	w.consumerGroup.Consume()
}

func (w *worker) Close() error {
	err := w.consumerGroup.Close()
	if err != nil {
		return errors.Wrap(err, "error closing kafka consumer group for worker")
	}
	if w.sidelineProducer != nil {
		err = w.sidelineProducer.Close()
		if err != nil {
			return errors.Wrap(err, "error closing kafka producer for worker")
		}
	}
	return nil
}

// Setup is specified to satisfy the sarama ConsumerGroupHandler interface
func (w *worker) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is specified to satisfy the sarama ConsumerGroupHandler interface
func (w *worker) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim processes messages from a partition. It has this particular signature
// in order to satisfy the sarama ConsumerGroupHandler interface
func (w *worker) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		w.handleMsg(session, msg)
	}
	return nil
}

// handleMsg processes a single message, and handles concerns such as retries and sidelining
func (w *worker) handleMsg(session sarama.ConsumerGroupSession, msg *sarama.ConsumerMessage) {
	defer session.MarkMessage(msg, "")

	protoMsg := w.taskHandler.PayloadProto()
	err := proto.Unmarshal(msg.Value, protoMsg)
	if err != nil {
		notifier.Notify(errors.Wrap(
			err,
			fmt.Sprintf("error parsing message as proto. Topic: %v, partition: %v, offset: %v",
				msg.Topic, msg.Partition, msg.Offset,
			),
		))
		return
	}

	ctx := context.Background()
	for attempt := 1; w.config.Retries < 0 || attempt <= 1+w.config.Retries; attempt++ {
		err := w.taskHandler.Handle(ctx, protoMsg)
		if err == nil {
			return
		}
		notifier.Notify(errors.Wrap(err, fmt.Sprintf("error in task handler for %v topic worker", w.config.Topic)))
		time.Sleep(w.config.RetryInterval)
	}

	if w.sidelineProducer == nil {
		return
	}
	err = w.sidelineProducer.Publish(ctx, w.config.Sideline.SidelineTopic, string(msg.Key), protoMsg)
	if err != nil {
		notifier.Notify(errors.Wrap(
			err,
			fmt.Sprintf("error producing failed message to sideline topic %v. Topic: %v, partition: %v, offset: %v",
				w.config.Sideline.SidelineTopic, msg.Topic, msg.Partition, msg.Offset,
			),
		))
	}
}

// NewWorker returns a worker which processes messages from a single topic.
// Messages which cannot be handled are retried the configured number of times.
// When all retries have failed, the message is pushed to a sideline topic, if
// specified.
func NewWorker(config Config, handler TaskHandler) (Worker, error) {
	w := &worker{
		config:      config,
		taskHandler: handler,
	}

	cg, err := kafka.NewConsumerGroup(newKafkaCGConfig(config), w)
	if err != nil {
		return nil, errors.Wrap(err, "error creating kafka consumer group for worker")
	}
	w.consumerGroup = cg

	if config.Sideline != nil {
		producer, err := kafka.NewProducer(newKafkaProducerConfig(config))
		if err != nil {
			return nil, errors.Wrap(err, "error creating kafka producer for worker")
		}
		w.sidelineProducer = producer
	}

	return w, nil
}
