package kafka

import (
	"time"

	"github.com/Shopify/sarama"
)

func newConfig() config {
	return config{
		saramaConfig: sarama.NewConfig(),
		errorHandler: defaultErrorHandler,
	}
}

type config struct {
	saramaConfig *sarama.Config
	errorHandler func(error)
}

// Option is used to configure the Kafka producer
type Option interface {
	apply(cfg *config)
}

type flushIntervalOption struct {
	interval time.Duration
}

func (opt *flushIntervalOption) apply(cfg *config) {
	cfg.saramaConfig.Producer.Flush.Frequency = opt.interval
}

// WithFlushInterval specifies a flush interval for the Kafka producer
func WithFlushInterval(interval time.Duration) Option {
	return &flushIntervalOption{
		interval: interval,
	}
}

type maxRetriesOption struct {
	maxRetries int
}

func (opt *maxRetriesOption) apply(cfg *config) {
	cfg.saramaConfig.Producer.Retry.Max = opt.maxRetries
}

// WithMaxRetries specifies the total number of times to retry sending a message
func WithMaxRetries(maxRetries int) Option {
	return &maxRetriesOption{
		maxRetries: maxRetries,
	}
}

type errorHandlerOption struct {
	errorHandler func(error)
}

func (opt *errorHandlerOption) apply(cfg *config) {
	cfg.errorHandler = opt.errorHandler
}

// WithErrorHandler specifies a custom error handler for the Kafka producer
func WithErrorHandler(errorHandler func(error)) Option {
	return &errorHandlerOption{
		errorHandler: errorHandler,
	}
}
