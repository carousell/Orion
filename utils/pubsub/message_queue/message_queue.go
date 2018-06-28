package message_queue

import (
	"log"
	"strconv"

	goPubSub "cloud.google.com/go/pubsub"
	cache "github.com/patrickmn/go-cache"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

type MessageQueue interface {
	Init(pubSubKey string, gProject string) error
	Close() error
	Publish(string, *PubSubData) *goPubSub.PublishResult
	GetResult(ctx context.Context, result *goPubSub.PublishResult) (string, error)
}

type PubSubData struct {
	Id        string
	Timestamp int64
	Data      []byte
}

type PubSubQueue struct {
	pubSubKey    string
	gProject     string
	PubsubClient *goPubSub.Client
	ctx          context.Context
	topics       *cache.Cache
}

func NewMessageQueue(enabled bool, serviceAccountKey string, project string) MessageQueue {
	MessageQueue := new(PubSubQueue)
	if enabled {
		MessageQueue.Init(serviceAccountKey, project)
	}
	return MessageQueue
}

func (pubsubqueue *PubSubQueue) Init(pubSubKey string, gProject string) error {
	var err error
	pubsubqueue.pubSubKey = pubSubKey
	pubsubqueue.gProject = gProject
	pubsubqueue.ctx, pubsubqueue.PubsubClient, err = pubsubqueue.configurePubsub()
	if err != nil {
		log.Fatalln("Error in client connections to PubSub", err)
		return err
	}
	pubsubqueue.topics = cache.New(cache.NoExpiration, cache.NoExpiration)
	return nil
}

func (pubsubqueue *PubSubQueue) Close() error {
	for _, item := range pubsubqueue.topics.Items() {
		if topic, ok := item.Object.(*goPubSub.Topic); ok {
			topic.Stop()
		}
	}
	return nil
}

func (pubsubqueue *PubSubQueue) configurePubsub() (context.Context, *goPubSub.Client, error) {
	var err error
	key := []byte(pubsubqueue.pubSubKey)
	conf, err := google.JWTConfigFromJSON(key, "https://www.googleapis.com/auth/pubsub")
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	ts := conf.TokenSource(ctx)
	ps, err := goPubSub.NewClient(ctx, pubsubqueue.gProject, option.WithTokenSource(ts))
	if err != nil {
		log.Fatal("Error in client connections to PubSub", err)
		return nil, nil, err
	}
	return ctx, ps, nil
}

func (pubsubqueue *PubSubQueue) Publish(topicName string, pubSubData *PubSubData) *goPubSub.PublishResult {
	var topic *goPubSub.Topic
	if t, ok := pubsubqueue.topics.Get(topicName); ok {
		if to, ok := t.(*goPubSub.Topic); ok {
			topic = to
		}
	}
	if topic == nil {
		topic = pubsubqueue.PubsubClient.Topic(topicName)
		pubsubqueue.topics.SetDefault(topicName, topic)
	}
	attributes := map[string]string{
		"id":        pubSubData.Id,
		"timestamp": strconv.FormatInt(pubSubData.Timestamp, 10),
	}
	publishResult := topic.Publish(pubsubqueue.ctx, &goPubSub.Message{Data: pubSubData.Data, Attributes: attributes})
	return publishResult
}

func (pubsubqueue *PubSubQueue) GetResult(ctx context.Context, result *goPubSub.PublishResult) (string, error) {
	return result.Get(ctx)
}
