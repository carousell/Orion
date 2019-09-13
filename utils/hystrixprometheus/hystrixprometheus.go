package hystrixprometheus

import (
	"github.com/afex/hystrix-go/hystrix/metric_collector"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

type PrometheusCollector struct {
	circuitOpen             *prometheus.GaugeVec
	attempts                *prometheus.CounterVec
	errors                  *prometheus.CounterVec
	successes               *prometheus.CounterVec
	failures                *prometheus.CounterVec
	rejects                 *prometheus.CounterVec
	shortCircuits           *prometheus.CounterVec
	timeouts                *prometheus.CounterVec
	fallbackSuccesses       *prometheus.CounterVec
	fallbackFailures        *prometheus.CounterVec
	contextCanceled         *prometheus.CounterVec
	contextDeadlineExceeded *prometheus.CounterVec
	totalDuration           *prometheus.HistogramVec
	runDuration             *prometheus.HistogramVec
}

func NewPrometheusCollector(namespace string, reg prometheus.Registerer, durationBuckets []float64) *PrometheusCollector {
	if namespace == "" {
		namespace = "hystrix"
	}

	if durationBuckets == nil {
		durationBuckets = prometheus.DefBuckets
	}

	pc := &PrometheusCollector{
		circuitOpen: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "circuit_open",
			Help:      "Status of the circuit. Zero value means a closed circuit.",
		}, []string{"command"}),
		attempts: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "attempts_total",
			Help:      "The number of requests.",
		}, []string{"command"}),
		errors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "errors_total",
			Help:      "The number of unsuccessful attempts. Attempts minus Errors will equal successes within a time range. Errors are any result from an attempt that is not a success.",
		}, []string{"command"}),
		successes: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "successes_total",
			Help:      "The number of requests that succeed.",
		}, []string{"command"}),
		failures: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "failures_total",
			Help:      "The number of requests that fail.",
		}, []string{"command"}),
		rejects: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "rejects_total",
			Help:      "The number of requests that are rejected.",
		}, []string{"command"}),
		shortCircuits: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "short_circuits_total",
			Help:      "The number of requests that short circuited due to the circuit being open.",
		}, []string{"command"}),
		timeouts: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "timeouts_total",
			Help:      "The number of requests that are timeouted in the circuit breaker.",
		}, []string{"command"}),
		fallbackSuccesses: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "fallback_successes_total",
			Help:      "The number of successes that occurred during the execution of the fallback function.",
		}, []string{"command"}),
		fallbackFailures: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "fallback_failures_total",
			Help:      "The number of failures that occurred during the execution of the fallback function.",
		}, []string{"command"}),
		contextCanceled: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "context_canceled_total",
			Help:      "The number of context canceled.",
		}, []string{"command"}),
		contextDeadlineExceeded: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "context_deadline_exceeded_total",
			Help:      "The number of context deadline exceeded.",
		}, []string{"command"}),
		totalDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "total_duration_seconds",
			Help:      "The total duration, includes thread queuing/scheduling/execution, semaphores, circuit breaker logic, and other aspects of overhead, of a Hystrix command.",
			Buckets:   durationBuckets,
		}, []string{"command"}),
		runDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "run_duration_seconds",
			Help:      "The duration of Hystrix command execution. This only measure the command only, without Hystrix overhead.",
			Buckets:   durationBuckets,
		}, []string{"command"}),
	}
	if reg != nil {
		reg.MustRegister(
			pc.circuitOpen,
			pc.attempts,
			pc.errors,
			pc.successes,
			pc.failures,
			pc.rejects,
			pc.shortCircuits,
			pc.timeouts,
			pc.fallbackSuccesses,
			pc.fallbackFailures,
			pc.contextCanceled,
			pc.contextDeadlineExceeded,
			pc.totalDuration,
			pc.runDuration,
		)
	} else {
		prometheus.MustRegister(
			pc.circuitOpen,
			pc.attempts,
			pc.errors,
			pc.successes,
			pc.failures,
			pc.rejects,
			pc.shortCircuits,
			pc.timeouts,
			pc.fallbackSuccesses,
			pc.fallbackFailures,
			pc.contextCanceled,
			pc.contextDeadlineExceeded,
			pc.totalDuration,
			pc.runDuration,
		)
	}
	return pc
}

type cmdCollector struct {
	commandName string
	metrics     *PrometheusCollector
}

func (pc *PrometheusCollector) Collector(name string) metricCollector.MetricCollector {
	c := &cmdCollector{
		commandName: name,
		metrics:     pc,
	}
	c.initCounters()
	return c
}

func (c *cmdCollector) initCounters() {
	c.metrics.circuitOpen.WithLabelValues(c.commandName).Set(0.0)
	c.metrics.attempts.WithLabelValues(c.commandName).Add(0.0)
	c.metrics.errors.WithLabelValues(c.commandName).Add(0.0)
	c.metrics.successes.WithLabelValues(c.commandName).Add(0.0)
	c.metrics.failures.WithLabelValues(c.commandName).Add(0.0)
	c.metrics.rejects.WithLabelValues(c.commandName).Add(0.0)
	c.metrics.shortCircuits.WithLabelValues(c.commandName).Add(0.0)
	c.metrics.timeouts.WithLabelValues(c.commandName).Add(0.0)
	c.metrics.fallbackSuccesses.WithLabelValues(c.commandName).Add(0.0)
	c.metrics.fallbackFailures.WithLabelValues(c.commandName).Add(0.0)
	c.metrics.contextCanceled.WithLabelValues(c.commandName).Add(0.0)
	c.metrics.contextDeadlineExceeded.WithLabelValues(c.commandName).Add(0.0)
}

func (c *cmdCollector) setGaugeMetric(pg *prometheus.GaugeVec, i float64) {
	pg.WithLabelValues(c.commandName).Set(i)
}

func (c *cmdCollector) incrementCounterMetric(pc *prometheus.CounterVec, i float64) {
	if i == 0 {
		return
	}
	pc.WithLabelValues(c.commandName).Add(i)
}

func (c *cmdCollector) updateTimerMetric(ph *prometheus.HistogramVec, dur time.Duration) {
	ph.WithLabelValues(c.commandName).Observe(dur.Seconds())
}

func (c *cmdCollector) Update(r metricCollector.MetricResult) {
	if r.Successes > 0 {
		c.setGaugeMetric(c.metrics.circuitOpen, 0)
	} else if r.ShortCircuits > 0 {
		c.setGaugeMetric(c.metrics.circuitOpen, 1)
	}

	c.incrementCounterMetric(c.metrics.attempts, r.Attempts)
	c.incrementCounterMetric(c.metrics.errors, r.Errors)
	c.incrementCounterMetric(c.metrics.successes, r.Successes)
	c.incrementCounterMetric(c.metrics.failures, r.Failures)
	c.incrementCounterMetric(c.metrics.rejects, r.Rejects)
	c.incrementCounterMetric(c.metrics.shortCircuits, r.ShortCircuits)
	c.incrementCounterMetric(c.metrics.timeouts, r.Timeouts)
	c.incrementCounterMetric(c.metrics.fallbackSuccesses, r.FallbackSuccesses)
	c.incrementCounterMetric(c.metrics.fallbackFailures, r.FallbackFailures)
	c.incrementCounterMetric(c.metrics.contextCanceled, r.ContextCanceled)
	c.incrementCounterMetric(c.metrics.contextDeadlineExceeded, r.ContextDeadlineExceeded)

	c.updateTimerMetric(c.metrics.totalDuration, r.TotalDuration)
	c.updateTimerMetric(c.metrics.runDuration, r.RunDuration)
}

// Reset is a noop operation in this collector.
func (c *cmdCollector) Reset() {}
