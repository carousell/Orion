package main

import (
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/carousell/go-utils/utils/errors/notifier"
	"github.com/carousell/healthcheck"
	"github.com/go-kit/kit/log"
	newrelic "github.com/newrelic/go-agent"
	stdopentracing "github.com/opentracing/opentracing-go"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
	"github.com/spf13/viper"
)

// initializes new relic, new relic key should be set to
// new-relic-api-key
func initNewRelic() (newrelic.Application, error) {
	apiKey := viper.GetString("new-relic.api-key")
	if strings.TrimSpace(apiKey) == "" {
		return nil, nil
	}
	config := newrelic.NewConfig(serviceName, apiKey)
	return newrelic.NewApplication(config)
}

func setupHealthcheck() {
	healthcheck.GetHealthCheck().SetHealth(true)
}

func initZipkin() stdopentracing.Tracer {
	zipkinAddr := viper.GetString(getServiceLower() + ".ZipkinAddr")
	var tracer stdopentracing.Tracer
	{
		if zipkinAddr != "" {
			logger := log.With(logger, "tracer", "Zipkin")
			logger.Log("addr", zipkinAddr)
			var collector zipkin.Collector
			var err error
			if strings.HasPrefix(zipkinAddr, "http") {
				collector, err = zipkin.NewHTTPCollector(
					zipkinAddr,
					zipkin.HTTPLogger(logger),
				)
				zipkin.HTTPBatchSize(200)
			} else {
				collector, err = zipkin.NewKafkaCollector(
					strings.Split(zipkinAddr, ","),
					zipkin.KafkaLogger(logger),
				)
			}
			if err != nil {
				logger.Log("err", err)
				os.Exit(1)
			}
			tracer, err = zipkin.NewTracer(
				zipkin.NewRecorder(collector, false, getHostname(), serviceName),
			)
			if err != nil {
				logger.Log("err", err)
				os.Exit(1)
			}
			stdopentracing.SetGlobalTracer(tracer)
		} else {
			logger := log.With(logger, "tracer", "none")
			logger.Log()
			tracer = stdopentracing.GlobalTracer() // no-op
		}
	}
	return tracer
}

func initHystrix() {
	hystrix.DefaultTimeout = 1000 // one sec
	hystrix.DefaultMaxConcurrent = 300
	hystrix.DefaultErrorPercentThreshold = 75
	hystrix.DefaultSleepWindow = 1000
	hystrix.DefaultVolumeThreshold = 75

	hystrixStreamHandler := hystrix.NewStreamHandler()
	hystrixStreamHandler.Start()
	port := viper.GetString(getServiceLower() + ".HystrixPort")
	logger.Log("HystrixPort", port)
	go http.ListenAndServe(net.JoinHostPort("", port), hystrixStreamHandler)
}

func initNotifier() {
	token := viper.GetString("rollbar.token")
	if strings.TrimSpace(token) == "" {
		return
	}
	env := viper.GetString(getServiceLower() + ".env")
	// environment for error notification
	notifier.SetEnvironemnt(env)
	// rollbar
	notifier.InitRollbar(token, env)
	logger.Log("token", token, "env", env)
}
