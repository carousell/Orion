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

const (
	pubsubTopic = "test_topic"
	pubsubMsg   = "test data"
	serverID    = "some serverId"
)

func setupMessageQueueMockCall() (*mockMessageQueue.MessageQueue, *goPubSub.PublishResult) {
	mockMessageQueue := &mockMessageQueue.MessageQueue{}
	newMessageQueueFn = func(enabled bool, serviceAccountKey string, project string) message_queue.MessageQueue {
		return mockMessageQueue
	}
	result := &goPubSub.PublishResult{}
	mockMessageQueue.On("Publish", pubsubTopic, mock.MatchedBy(func(pubsubData *message_queue.PubSubData) bool {
		return pubsubMsg == string(pubsubData.Data)
	})).Return(result)
	return mockMessageQueue, result
}

func TestPublishMessageSync(t *testing.T) {
	ctx := context.Background()
	mockMessageQueue, result := setupMessageQueueMockCall()
	mockMessageQueue.On("GetResult", ctx, result).Return(serverID, nil)

	p := NewPubSubService(PubSubConfig{})
	data := []byte(pubsubMsg)
	response, err := p.PublishMessage(ctx, pubsubTopic, data, true)
	assert.Nil(t, response)
	assert.Nil(t, err)
	p.Close()

	call := mockMessageQueue.Mock.ExpectedCalls[0]
	assert.Equal(t, "Publish", call.Method)
}

func TestPublishMessageAsync(t *testing.T) {
	ctx := context.Background()
	mockMessageQueue, result := setupMessageQueueMockCall()

	p := NewPubSubService(PubSubConfig{})
	data := []byte(pubsubMsg)
	response, err := p.PublishMessage(ctx, pubsubTopic, data, false)
	assert.Equal(t, result, response)
	assert.Nil(t, err)
	p.Close()

	call := mockMessageQueue.Mock.ExpectedCalls[0]
	assert.Equal(t, "Publish", call.Method)
}
