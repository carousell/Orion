package pubsub

import (
	"context"
	"github.com/carousell/Orion/utils/log/loggers"
	"time"

	goPubSub "cloud.google.com/go/pubsub"
	"github.com/afex/hystrix-go/hystrix"
	"github.com/carousell/Orion/utils/executor"
	"github.com/carousell/Orion/utils/log"
	messageQueue "github.com/carousell/Orion/utils/pubsub/message_queue"
	"github.com/carousell/Orion/utils/spanutils"
)

//Config is the config for pubsub
type Config struct {
	Key                    string
	Project                string
	Enabled                bool
	Timeout                int
	BulkPublishConcurrency int
	Retries                int
}

//Service is the interface implemented by a pubsub service
type Service interface {
	PublishMessage(ctx context.Context, topic string, data []byte, waitSync bool) (*goPubSub.PublishResult, error)
	BulkPublishMessages(ctx context.Context, topic string, data [][]byte, waitSync bool)
	SubscribeMessages(ctx context.Context, subscribe string, subscribeFunction messageQueue.SubscribeFunction) error
	Close()
}

type pubSubService struct {
	MessageQueue messageQueue.MessageQueue
	Config       Config
}

var newMessageQueueFn = messageQueue.NewMessageQueue
var newExecutorFn = executor.NewExecutor

//NewPubSubService build and returns an pubsub service handler
func NewPubSubService(config Config) Service {
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
			log.Error(ctx, "error in pubsub msg publish", []loggers.Label{{"err", er}, {"component", "pubsub PublishMessage"}})
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

//SubscribeMessages Subscirbes to pubsub and returns error if any
func (g *pubSubService) SubscribeMessages(ctx context.Context, subscribe string, subscribeFunction messageQueue.SubscribeFunction) error {
	return g.MessageQueue.SubscribeMessages(ctx, subscribe, subscribeFunction)
}
