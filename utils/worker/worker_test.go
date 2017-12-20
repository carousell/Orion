package worker

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLocalWoker(t *testing.T) {
	cfg := Config{}
	cfg.LocalMode = true
	executed := false
	w := NewWorker(cfg)
	w.RegisterTask("test", func(ctx context.Context, payload string) error {
		executed = true
		return nil
	})
	w.RunWorker("hello", 1)
	w.Schedule(context.Background(), "test", "hello")
	time.Sleep(time.Second * 1)
	assert.True(t, executed, "should execute task")
}

func TestLocalWokerFail(t *testing.T) {
	cfg := Config{}
	cfg.LocalMode = true
	executed := false
	w := NewWorker(cfg)
	w.RegisterTask("test", func(ctx context.Context, payload string) error {
		executed = true
		return nil
	})
	w.RunWorker("hello", 1)
	w.Schedule(context.Background(), "test2", "hello")
	time.Sleep(time.Second * 1)
	assert.False(t, executed, "should execute task")
}

/*
func TestRemoteWorker(t *testing.T) {
	cfg := Config{}
	cfg.LocalMode = false
	cfg.RabbitConfig = new(RabbitMQConfig)
	cfg.RabbitConfig.Host = "192.168.99.100"
	cfg.RabbitConfig.QueueName = "test"
	cfg.RabbitConfig.UserName = "guest"
	cfg.RabbitConfig.Password = "guest"
	executed := false
	w := NewWorker(cfg)
	w.RegisterTask("test", func(ctx context.Context, payload string) error {
		fmt.Println("executing", payload)
		executed = true
		return nil
	})
	w.RunWorker("hello", 1)
	for i := 0; i < 100; i++ {
		assert.NoError(t, w.Schedule(context.Background(), "test", "hello"), "scheduling task resulted in error")
		time.Sleep(time.Millisecond * 50)
	}
	time.Sleep(time.Second * 5)
	assert.True(t, executed, "should execute task")
}
*/
