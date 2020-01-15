package kafka

import (
	"strings"
	"testing"

	"context"

	"github.com/Shopify/sarama"
	"github.com/Shopify/sarama/mocks"
	"github.com/carousell/Orion/utils/errors"
)

func TestProducer(t *testing.T) {

	cfg := sarama.Config{}
	cfg.Producer.Return.Successes = true
	cfg.Producer.Return.Errors = true
	ctx := context.Background()

	t.Run("open producer accepts messages", func(t *testing.T) {
		asyncProducer := mocks.NewAsyncProducer(t, &cfg)
		asyncProducer.ExpectInputAndSucceed()
		producer := Producer{
			asyncProducer: asyncProducer,
		}
		producer.Run()
		defer producer.Close()

		err := producer.Produce(ctx, "test-topic", "test-key", []byte("test-payload"))
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		resp := <-asyncProducer.Successes()
		val, _ := resp.Value.Encode()
		if string(val) != "test-payload" {
			t.Fatalf("expected value sent to Kafka to be %v, got %v", "test-payload", string(val))
		}
	})

	t.Run("producer errors are handled by custom handler", func(t *testing.T) {
		asyncProducer := mocks.NewAsyncProducer(t, &cfg)
		testErr := errors.New("test error")
		asyncProducer.ExpectInputAndFail(testErr)

		errCh := make(chan error)
		errorHandler := func(err error) {
			errCh <- err
		}

		producer := Producer{
			asyncProducer: asyncProducer,
			errorHandler:  errorHandler,
		}
		producer.Run()
		defer producer.Close()

		err := producer.Produce(ctx, "test-topic", "test-key", []byte("test-payload"))
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		producerErr := <-errCh
		if !strings.Contains(producerErr.Error(), testErr.Error()) {
			t.Fatalf("expected error handler to receive %v, got %v", testErr, producerErr)
		}
	})

	t.Run("uninitialized producer does not accept messages", func(t *testing.T) {
		producer := Producer{}
		err := producer.Produce(ctx, "test-topic", "test-key", []byte("test-payload"))
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("closed producer does not accept messages", func(t *testing.T) {
		asyncProducer := mocks.NewAsyncProducer(t, &cfg)
		producer := Producer{
			asyncProducer: asyncProducer,
		}
		producer.Run()
		producer.Close()

		err := producer.Produce(ctx, "test-topic", "test-key", []byte("test-payload"))
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
