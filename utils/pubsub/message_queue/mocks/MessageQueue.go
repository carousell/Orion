// Code generated by mockery v1.0.0. DO NOT EDIT.
package mocks

import context "context"
import message_queue "github.com/carousell/Orion/utils/pubsub/message_queue"
import mock "github.com/stretchr/testify/mock"
import pubsub "cloud.google.com/go/pubsub"

// MessageQueue is an autogenerated mock type for the MessageQueue type
type MessageQueue struct {
	mock.Mock
}

// Close provides a mock function with given fields:
func (_m *MessageQueue) Close() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetResult provides a mock function with given fields: ctx, result
func (_m *MessageQueue) GetResult(ctx context.Context, result *pubsub.PublishResult) (string, error) {
	ret := _m.Called(ctx, result)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, *pubsub.PublishResult) string); ok {
		r0 = rf(ctx, result)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *pubsub.PublishResult) error); ok {
		r1 = rf(ctx, result)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Init provides a mock function with given fields: pubSubKey, gProject
func (_m *MessageQueue) Init(pubSubKey string, gProject string) error {
	ret := _m.Called(pubSubKey, gProject)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(pubSubKey, gProject)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Publish provides a mock function with given fields: _a0, _a1
func (_m *MessageQueue) Publish(_a0 string, _a1 *message_queue.PubSubData) *pubsub.PublishResult {
	ret := _m.Called(_a0, _a1)

	var r0 *pubsub.PublishResult
	if rf, ok := ret.Get(0).(func(string, *message_queue.PubSubData) *pubsub.PublishResult); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*pubsub.PublishResult)
		}
	}

	return r0
}

// SubscribeMessages provides a mock function with given fields: ctx, subscriptionId, subscribeFunction
func (_m *MessageQueue) SubscribeMessages(ctx context.Context, subscriptionId string, subscribeFunction message_queue.SubscribeFunction) error {
	ret := _m.Called(ctx, subscriptionId, subscribeFunction)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, message_queue.SubscribeFunction) error); ok {
		r0 = rf(ctx, subscriptionId, subscribeFunction)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
