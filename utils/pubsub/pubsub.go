package pubsub

import (
	"context"
	"time"

	goPubSub "cloud.google.com/go/pubsub"
	"github.com/afex/hystrix-go/hystrix"
	"github.com/carousell/Orion/utils/executor"
	messageQueue "github.com/carousell/Orion/utils/pubsub/message_queue"
	"github.com/carousell/Orion/utils/spanutils"
)

type PubSubConfig struct {
	Key                    string
	Project                string
	Enabled                bool
	Timeout                int
	BulkPublishConcurrency int
}

type PubSubService interface {
	PublishMessage(ctx context.Context, topic string, data []byte, waitSync bool) *goPubSub.PublishResult
	BulkPublishMessages(ctx context.Context, topic string, data [][]byte)
	Close()
}

type pubSubService struct {
	MessageQueue messageQueue.MessageQueue
	Config       PubSubConfig
}

//NewPubSubService build and returns an pubsub service handler
func NewPubSubService(config PubSubConfig) PubSubService {
	MessageQueue := new(messageQueue.PubSubQueue)
	if config.Enabled {
		MessageQueue.Init(config.Key, config.Project)
	}
	hysConfig := hystrix.CommandConfig{Timeout: config.Timeout}
	hystrix.ConfigureCommand("PubSubPublish", hysConfig)
	return &pubSubService{
		MessageQueue: MessageQueue,
		Config:       config,
	}
}

//Close closes any active connection to Pubsub endpoint
func (g *pubSubService) Close() {
	if g.Config.Enabled {
		g.MessageQueue.Close()
	}
}

//PublishMessage publishes a single message to give topic, set waitSync param to true to wait for publish ack
func (g *pubSubService) PublishMessage(ctx context.Context, topic string, data []byte, waitSync bool) *goPubSub.PublishResult {
	var result *goPubSub.PublishResult
	hystrix.Do("PubSubPublish", func() error {
		span, _ := spanutils.NewExternalSpan(ctx, "PubSubPublish", topic)
		// zipkin span
		defer span.Finish()
		pubsubData := new(messageQueue.PubSubData)
		pubsubData.Data = data
		pubsubData.Timestamp = time.Now().UnixNano() / 1000000
		result = g.MessageQueue.Publish(topic, pubsubData)
		if waitSync {
			result.Get(ctx)
		}
		return nil
	}, nil)
	return result
}

//BulkPublishMessages publishes a multiple message to give topic, with "BulkPublishConcurrency" no of routines
func (g *pubSubService) BulkPublishMessages(ctx context.Context, topic string, data [][]byte) {
	e := executor.NewExecutor(executor.WithFailOnError(false), executor.WithConcurrency(g.Config.BulkPublishConcurrency))
	for _, v := range data {
		singleMsg := v
		e.Add(func() error {
			_ = g.PublishMessage(ctx, topic, singleMsg, true)
			return nil
		})
	}
	e.Wait()
}
