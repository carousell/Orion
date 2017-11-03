package orion

import (
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/afex/hystrix-go/hystrix"
	logg "github.com/go-kit/kit/log"
	newrelic "github.com/newrelic/go-agent"
	stdopentracing "github.com/opentracing/opentracing-go"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
)

//InitHystrix initializes hystrix with default values
func (d *DefaultServerImpl) InitHystrix() {
	hystrix.DefaultTimeout = 1000 // one sec
	hystrix.DefaultMaxConcurrent = 300
	hystrix.DefaultErrorPercentThreshold = 75
	hystrix.DefaultSleepWindow = 1000
	hystrix.DefaultVolumeThreshold = 75

	hystrixStreamHandler := hystrix.NewStreamHandler()
	hystrixStreamHandler.Start()
	port := d.GetOrionConfig().HystrixConfig.Port
	log.Println("HystrixPort", port)
	go http.ListenAndServe(net.JoinHostPort("", port), hystrixStreamHandler)
}

//InitZipkin initializes zipkin collectors and traces to default config
func (d *DefaultServerImpl) InitZipkin() {
	zipkinAddr := d.GetOrionConfig().ZipkinConfig.Addr
	serviceName := d.GetOrionConfig().OrionServerName
	if zipkinAddr != "" {
		logger := logg.NewLogfmtLogger(os.Stdout)
		logger = logg.With(logger, "ts", logg.DefaultTimestampUTC)
		logger.Log("zipkin-addr", zipkinAddr)
		var collector zipkin.Collector
		var err error
		if strings.HasPrefix(zipkinAddr, "http") {
			collector, err = zipkin.NewHTTPCollector(
				zipkinAddr,
				zipkin.HTTPLogger(logger),
			)
			zipkin.HTTPBatchSize(1)
		} else {
			collector, err = zipkin.NewKafkaCollector(
				strings.Split(zipkinAddr, ","),
				zipkin.KafkaLogger(logger),
			)
		}
		if err != nil {
			logger.Log("err", err)
		}

		tracer, err := zipkin.NewTracer(
			zipkin.NewRecorder(collector, true, getHostname(), serviceName),
		)
		if err != nil {
			logger.Log("err", err)
		} else {
			stdopentracing.SetGlobalTracer(tracer)
		}
	} else {
		stdopentracing.SetGlobalTracer(stdopentracing.NoopTracer{})
	}
}

//InitNewRelic initializes newrelic lib to default values based on config
func (d *DefaultServerImpl) InitNewRelic() {
	apiKey := d.GetOrionConfig().NewRelicConfig.APIKey
	if strings.TrimSpace(apiKey) == "" {
		return
	}
	serviceName := d.GetOrionConfig().NewRelicConfig.ServiceName
	if strings.TrimSpace(serviceName) == "" {
		serviceName = d.GetOrionConfig().OrionServerName
	}
	config := newrelic.NewConfig(serviceName, apiKey)
	app, err := newrelic.NewApplication(config)
	if err != nil {
		log.Println("nr-error", err)
	} else {
		log.Println("NR", "initialized with "+serviceName)
		d.nrApp = app
	}
}

func initInitializers(d interface{}) {

	// pre init
	if i, ok := d.(PreInitializer); ok {
		i.PreInit()
	}

	// process hystrix
	if i, ok := d.(HystrixInitializer); ok {
		i.InitHystrix()
	}

	// process zipkin
	if i, ok := d.(ZipkinInitializer); ok {
		i.InitZipkin()
	}

	// process newrelic
	if i, ok := d.(NewRelicInitializer); ok {
		i.InitNewRelic()
	}

	// post init
	if i, ok := d.(PostInitializer); ok {
		i.PostInit()
	}
}
