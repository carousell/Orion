package pubsub

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	goPubSub "cloud.google.com/go/pubsub"
	"github.com/carousell/Orion/utils/executor"
	"github.com/carousell/Orion/utils/pubsub/message_queue"
	mockMessageQueue "github.com/carousell/Orion/utils/pubsub/message_queue/mocks"
)

const (
	pubsubTopic = "test_topic"
	pubsubMsg   = "test data"
	serverID    = "some serverId"
)

func setupMessageQueueMockCall(publishMethodCallCount int) (*mockMessageQueue.MessageQueue, *goPubSub.PublishResult) {
	mockMessageQueue := &mockMessageQueue.MessageQueue{}
	newMessageQueueFn = func(enabled bool, serviceAccountKey string, project string) message_queue.MessageQueue {
		return mockMessageQueue
	}
	result := &goPubSub.PublishResult{}
	mockMessageQueue.On("Publish", pubsubTopic, mock.MatchedBy(func(pubsubData *message_queue.PubSubData) bool {
		return pubsubMsg == string(pubsubData.Data)
	})).Return(result).Times(publishMethodCallCount)
	return mockMessageQueue, result
}

func TestPublishMessageSync(t *testing.T) {
	ctx := context.Background()
	mockMessageQueue, result := setupMessageQueueMockCall(1)
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
	mockMessageQueue, result := setupMessageQueueMockCall(1)

	p := NewPubSubService(PubSubConfig{})
	data := []byte(pubsubMsg)
	response, err := p.PublishMessage(ctx, pubsubTopic, data, false)
	assert.Equal(t, result, response)
	assert.Nil(t, err)
	p.Close()

	call := mockMessageQueue.Mock.ExpectedCalls[0]
	assert.Equal(t, "Publish", call.Method)
}

type mockExe struct {
	tasks []executor.Task
}

func (e *mockExe) Add(task executor.Task) {
	e.tasks = append(e.tasks, task)
}
func (e *mockExe) Wait() error {
	return nil
}

func TestBulkPublishMessageAsync(t *testing.T) {
	noOfDataToPublish := 2
	ctx := context.Background()
	mockMessageQueue, _ := setupMessageQueueMockCall(noOfDataToPublish)

	mockExecutor := new(mockExe)
	newExecutorFn = func(options ...executor.Option) executor.Executor {
		return mockExecutor
	}
	p := NewPubSubService(PubSubConfig{})
	data := [][]byte{[]byte(pubsubMsg), []byte(pubsubMsg)}
	p.BulkPublishMessages(ctx, pubsubTopic, data, false)
	p.Close()

	assert.Equal(t, noOfDataToPublish, len(mockExecutor.tasks))
	for i := 0; i < noOfDataToPublish; i++ {
		mockExecutor.tasks[i]()
	}

	call := mockMessageQueue.Mock.ExpectedCalls[0]
	assert.Equal(t, "Publish", call.Method)
}

func TestBulkPublishMessageSync(t *testing.T) {
	noOfDataToPublish := 2
	ctx := context.Background()
	mockMessageQueue, result := setupMessageQueueMockCall(noOfDataToPublish)
	mockMessageQueue.On("GetResult", ctx, result).Return(serverID, nil).Times(noOfDataToPublish)

	mockExecutor := new(mockExe)
	newExecutorFn = func(options ...executor.Option) executor.Executor {
		return mockExecutor
	}
	p := NewPubSubService(PubSubConfig{})
	data := [][]byte{[]byte(pubsubMsg), []byte(pubsubMsg)}
	p.BulkPublishMessages(ctx, pubsubTopic, data, true)
	p.Close()

	assert.Equal(t, noOfDataToPublish, len(mockExecutor.tasks))
	for i := 0; i < noOfDataToPublish; i++ {
		mockExecutor.tasks[i]()
	}

	call := mockMessageQueue.Mock.ExpectedCalls[0]
	assert.Equal(t, "Publish", call.Method)
}

type mockMessageQueueForRetry struct {
	tries int
}

func (_m *mockMessageQueueForRetry) Close() error {
	return nil
}
func (_m *mockMessageQueueForRetry) GetResult(ctx context.Context, result *goPubSub.PublishResult) (string, error) {
	_m.tries++
	return "", errors.New("Timeout")
}
func (_m *mockMessageQueueForRetry) Init(pubSubKey string, gProject string) error {
	_m.tries = 0
	return nil
}
func (_m *mockMessageQueueForRetry) Publish(_a0 string, _a1 *message_queue.PubSubData) *goPubSub.PublishResult {
	return nil
}
func (_m *mockMessageQueueForRetry) SubscribeMessages(ctx context.Context, subscriptionId string, subscribeFunction message_queue.SubscribeFunction) error {
	return nil
}

func TestPublishMessageSyncWithRetries(t *testing.T) {
	ctx := context.Background()
	mockMessageQueue := &mockMessageQueueForRetry{}
	newMessageQueueFn = func(enabled bool, serviceAccountKey string, project string) message_queue.MessageQueue {
		return mockMessageQueue
	}

	p := NewPubSubService(PubSubConfig{Retries: 2})
	data := []byte(pubsubMsg)
	response, err := p.PublishMessage(ctx, pubsubTopic, data, true)
	assert.Nil(t, response)
	assert.Nil(t, err)
	p.Close()

	// 1 + 2 reties
	assert.Equal(t, 3, mockMessageQueue.tries)
}

func testSubscriberFn(ctx context.Context, msg *goPubSub.Message) {
}
func TestSubscribeMessages(t *testing.T) {
	ctx := context.Background()
	mockMessageQueue := &mockMessageQueue.MessageQueue{}
	newMessageQueueFn = func(enabled bool, serviceAccountKey string, project string) message_queue.MessageQueue {
		return mockMessageQueue
	}
	mockMessageQueue.On("SubscribeMessages", ctx, "subscriptionId", mock.MatchedBy(func(subscriberFn message_queue.SubscribeFunction) bool {
		return true
	})).Return(nil)
	p := NewPubSubService(PubSubConfig{})
	p.SubscribeMessages(ctx, "subscriptionId", testSubscriberFn)

	call := mockMessageQueue.Mock.ExpectedCalls[0]
	assert.Equal(t, "SubscribeMessages", call.Method)
}
