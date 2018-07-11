package pubsub

import (
	"context"
	"log"
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
	Retries                int
}

type PubSubService interface {
	PublishMessage(ctx context.Context, topic string, data []byte, waitSync bool) (*goPubSub.PublishResult, error)
	BulkPublishMessages(ctx context.Context, topic string, data [][]byte, waitSync bool)

	SubscribeMessages(ctx context.Context, subscribe string, retryOnError bool) (SubscriberData, error)

	Close()
}

type pubSubService struct {
	MessageQueue messageQueue.MessageQueue
	Config       PubSubConfig
}

type SubscriberData struct {
	Data  chan goPubSub.Message
	Error chan error
}

var newMessageQueueFn = messageQueue.NewMessageQueue
var newExecutorFn = executor.NewExecutor

//NewPubSubService build and returns an pubsub service handler
func NewPubSubService(config PubSubConfig) PubSubService {
	hysConfig := hystrix.CommandConfig{Timeout: config.Timeout}
	hystrix.ConfigureCommand("PubSubPublish", hysConfig)
	return &pubSubService{
		MessageQueue: newMessageQueueFn(config.Enabled, config.Key, config.Project),
		Config:       config,
	}
}

//Close closes any active connection to Pubsub endpoint
func (g *pubSubService) Close() {
	if g.Config.Enabled {
		g.MessageQueue.Close()
	}
}

//Defaults to 1 retry
func (g *pubSubService) GetRetries() int {
	if g.Config.Retries < 1 {
		return 1
	}
	return g.Config.Retries
}

//PublishMessage publishes a single message to give topic, set waitSync param to true to wait for publish ack
func (g *pubSubService) PublishMessage(ctx context.Context, topic string, data []byte, waitSync bool) (*goPubSub.PublishResult, error) {
	var result *goPubSub.PublishResult
	retries := g.GetRetries()
	for retries >= 0 {
		retries--
		er := hystrix.Do("PubSubPublish", func() error {
			span, _ := spanutils.NewExternalSpan(ctx, "PubSubPublish", "/"+topic)
			// zipkin span
			defer span.Finish()
			pubsubData := new(messageQueue.PubSubData)
			pubsubData.Data = data
			pubsubData.Timestamp = time.Now().UnixNano() / 1000000
			result = g.MessageQueue.Publish(topic, pubsubData)
			if waitSync {
				_, err := g.MessageQueue.GetResult(ctx, result)
				if err != nil {
					return err
				}
			}
			return nil
		}, nil)
		if er != nil {
			log.Println("Error:", er.Error())
		} else {
			break
		}
	}
	if !waitSync {
		return result, nil
	}
	return nil, nil
}

//BulkPublishMessages publishes a multiple message to give topic, with "BulkPublishConcurrency" no of routines
func (g *pubSubService) BulkPublishMessages(ctx context.Context, topic string, data [][]byte, waitSync bool) {
	e := newExecutorFn(executor.WithFailOnError(false), executor.WithConcurrency(g.Config.BulkPublishConcurrency))
	for _, v := range data {
		singleMsg := v
		e.Add(func() error {
			_, err := g.PublishMessage(ctx, topic, singleMsg, waitSync)
			return err
		})
	}
	e.Wait()
}

//SubscribeMessages Subscirbes to pubsub and returns the received messages through the channel
func (g *pubSubService) SubscribeMessages(ctx context.Context, subscribe string, retryOnError bool) (SubscriberData, error) {
	subscriberData := SubscriberData{}
	subscriberData.Data, subscriberData.Error = g.MessageQueue.SubscribeMessages(ctx, subscribe)
	if retryOnError {
		go func() {
			for err := range subscriberData.Error {
				if err != nil {
					subscriberData.Data, subscriberData.Error = g.MessageQueue.SubscribeMessages(ctx, subscribe)
				}
			}
		}()
	}
	return subscriberData, nil
}
