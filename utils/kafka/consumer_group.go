package kafka

import (
	"context"

	"github.com/Shopify/sarama"
	"github.com/carousell/Orion/utils/errors"
	"github.com/carousell/Orion/utils/errors/notifier"
	"github.com/carousell/Orion/utils/log"
)

type ConsumerGroup interface {
	Consume()
	Close() error
}

type consumerGroup struct {
	consumerGroup        sarama.ConsumerGroup
	consumerGroupHandler sarama.ConsumerGroupHandler
	topic                string
	done                 chan bool
}

func NewConsumerGroup(config ConsumerGroupConfig, cgHandler sarama.ConsumerGroupHandler) (ConsumerGroup, error) {
	defaultConfig := sarama.NewConfig()
	defaultConfig.ClientID = config.ClientID
	defaultConfig.Version = config.KafkaVersion
	defaultConfig.Consumer.Offsets.CommitInterval = config.CommitInterval
	defaultConfig.Net.DialTimeout = config.KafkaSocketTimeout
	defaultConfig.Net.ReadTimeout = config.KafkaSocketTimeout
	defaultConfig.Net.WriteTimeout = config.KafkaSocketTimeout
	cg, err := sarama.NewConsumerGroup(config.Brokers, config.Name, defaultConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new kafka consumer group from client")
	}
	consumerGroup := &consumerGroup{
		consumerGroup:        cg,
		consumerGroupHandler: cgHandler,
		topic:                config.Topic,
		done:                 make(chan bool),
	}
	return consumerGroup, nil
}

func (c *consumerGroup) Consume() {
	ctx := context.Background()
	for {
		select {
		case <-c.done:
			return
		default:
			log.Info(ctx, "msg", "waiting to be assigned partitions", "topic", c.topic)
			err := c.consumerGroup.Consume(ctx, []string{c.topic}, c.consumerGroupHandler)
			if err != nil {
				wrappedErr := errors.Wrap(err, "error consuming topic "+c.topic)
				log.Error(ctx, "err", wrappedErr)
				notifier.Notify(wrappedErr)
			}
		}
	}
}

func (c *consumerGroup) Close() error {
	c.done <- true
	if c.consumerGroup == nil {
		return nil
	}
	return errors.Wrap(c.consumerGroup.Close(), "failed to close kafka consumer group")
}
