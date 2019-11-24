package kafka

import (
	"context"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"github.com/golang/protobuf/proto"
)

func TestKafkaProductionConsumption(t *testing.T) {
	// new producer
	producer, err := NewProducer(ProducerConfig{
		Brokers:            []string{"localhost:9092"},
		KafkaSocketTimeout: 2 * time.Second,
		ClientID:           "sarama",
		KafkaVersion:       sarama.V0_11_0_2,
		RequiredAcks:       sarama.WaitForAll,
	})
	if err != nil {
		t.FailNow()
	}

	// new consumer group
	handler := &sampleHandler{
		consumed: make(chan bool),
	}
	cg, err := NewConsumerGroup(ConsumerGroupConfig{
		Brokers:            []string{"localhost:9092"},
		KafkaSocketTimeout: 2 * time.Second,
		ClientID:           "sarama",
		KafkaVersion:       sarama.V0_11_0_2,
		Name:               "sarama.cg",
		Topic:              "sarama",
		CommitInterval:     time.Second,
	}, handler)
	if err != nil {
		t.FailNow()
	}

	// start consumer
	go cg.Consume()
	// wait for consumer to start
	time.Sleep(5 * time.Second)

	// produce
	err = producer.Publish(context.Background(), "sarama", "key", &Sample{
		Test: "test",
	})
	if err != nil {
		t.FailNow()
	}

	consumed := <-handler.consumed
	if !consumed {
		t.FailNow()
	}
}

type sampleHandler struct {
	consumed chan bool
}

func (h *sampleHandler) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h *sampleHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h *sampleHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		protoMsg := Sample{}
		err := proto.Unmarshal(msg.Value, &protoMsg)
		if err != nil {
			return err
		}
		session.MarkMessage(msg, "")
		if protoMsg.Test == "test" {
			h.consumed <- true
		}
	}
	return nil
}
