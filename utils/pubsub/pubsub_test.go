package pubsub

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"

	goPubSub "cloud.google.com/go/pubsub"
	"github.com/carousell/Orion/utils/pubsub/message_queue"
	mockMessageQueue "github.com/carousell/Orion/utils/pubsub/message_queue/mocks"
	"github.com/stretchr/testify/assert"
)

func TestPublishMessageSync(t *testing.T) {
	// defer leaktest.Check(t)()
	testConf := PubSubConfig{}
	mockMessageQueue := &mockMessageQueue.MessageQueue{}
	newMessageQueueFn = func(enabled bool, serviceAccountKey string, project string) message_queue.MessageQueue {
		return mockMessageQueue
	}

	result := &goPubSub.PublishResult{}
	// mockMessageQueue.On("Publish", "test_topic", mock.AnythingOfType("*message_queue.PubSubData")).Return(result)
	mockMessageQueue.On("Publish", "test_topic", mock.MatchedBy(func(pubsubData *message_queue.PubSubData) bool {
		return "test data" == string(pubsubData.Data)
	})).Return(result)

	p := NewPubSubService(testConf)
	data := []byte("test data")
	ctx := context.Background()
	_ = p.PublishMessage(ctx, "test_topic", data, true)
	p.Close()

	call := mockMessageQueue.Mock.ExpectedCalls[0]
	assert.Equal(t, "Publish", call.Method)
}
