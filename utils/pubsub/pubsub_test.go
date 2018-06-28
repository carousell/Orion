package pubsub

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	goPubSub "cloud.google.com/go/pubsub"
	"github.com/carousell/Orion/utils/pubsub/message_queue"
	mockMessageQueue "github.com/carousell/Orion/utils/pubsub/message_queue/mocks"
)

func TestPublishMessageSync(t *testing.T) {
	// defer leaktest.Check(t)()
	pubsubTopic := "test_topic"
	pubsubMsg := "test data"

	testConf := PubSubConfig{}
	mockMessageQueue := &mockMessageQueue.MessageQueue{}
	newMessageQueueFn = func(enabled bool, serviceAccountKey string, project string) message_queue.MessageQueue {
		return mockMessageQueue
	}

	result := &goPubSub.PublishResult{}
	mockMessageQueue.On("Publish", pubsubTopic, mock.MatchedBy(func(pubsubData *message_queue.PubSubData) bool {
		return pubsubMsg == string(pubsubData.Data)
	})).Return(result)

	p := NewPubSubService(testConf)
	data := []byte(pubsubMsg)
	ctx := context.Background()
	_ = p.PublishMessage(ctx, pubsubTopic, data, true)
	p.Close()

	call := mockMessageQueue.Mock.ExpectedCalls[0]
	assert.Equal(t, "Publish", call.Method)
}
