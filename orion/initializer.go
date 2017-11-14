package orion

import (
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/carousell/Orion/utils"
	logg "github.com/go-kit/kit/log"
	newrelic "github.com/newrelic/go-agent"
	stdopentracing "github.com/opentracing/opentracing-go"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
)

const (
	NR_APP = "INIT:NR_APP"
)

func DefaultInitializers() []Initializer {
	return []Initializer{
		HystrixInitializer(),
		ZipkinInitializer(),
		NewRelicInitializer(),
	}
}

func HystrixInitializer() Initializer {
	return &hystrixInitializer{}
}

func ZipkinInitializer() Initializer {
	return &zipkinInitializer{}
}

func NewRelicInitializer() Initializer {
	return &newRelicInitializer{}
}

type hystrixInitializer struct {
}

func (h *hystrixInitializer) Init(svr Server) error {
	hystrix.DefaultTimeout = 1000 // one sec
	hystrix.DefaultMaxConcurrent = 300
	hystrix.DefaultErrorPercentThreshold = 75
	hystrix.DefaultSleepWindow = 1000
	hystrix.DefaultVolumeThreshold = 75

	hystrixStreamHandler := hystrix.NewStreamHandler()
	hystrixStreamHandler.Start()
	port := svr.GetOrionConfig().HystrixConfig.Port
	log.Println("HystrixPort", port)
	go http.ListenAndServe(net.JoinHostPort("", port), hystrixStreamHandler)
	return nil
}

func (h *hystrixInitializer) ReInit(svr Server) error {
	// do nothing, cant be reinited
	return nil
}

type newRelicInitializer struct {
}

func (n *newRelicInitializer) Init(svr Server) error {
	apiKey := svr.GetOrionConfig().NewRelicConfig.APIKey
	if strings.TrimSpace(apiKey) == "" {
		return errors.New("empty token")
	}
	serviceName := svr.GetOrionConfig().NewRelicConfig.ServiceName
	if strings.TrimSpace(serviceName) == "" {
		serviceName = svr.GetOrionConfig().OrionServerName
	}
	config := newrelic.NewConfig(serviceName, apiKey)
	app, err := newrelic.NewApplication(config)
	if err != nil {
		log.Println("nr-error", err)
		return err
	} else {
		log.Println("NR", "initialized with "+serviceName)
		svr.Store(NR_APP, app)
	}
	return nil
}

func (n *newRelicInitializer) ReInit(svr Server) error {
	return n.Init(svr)
}

type zipkinInitializer struct {
}

func (z *zipkinInitializer) Init(svr Server) error {
	zipkinAddr := svr.GetOrionConfig().ZipkinConfig.Addr
	serviceName := svr.GetOrionConfig().OrionServerName
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
			return err
		}

		tracer, err := zipkin.NewTracer(
			zipkin.NewRecorder(collector, true, utils.GetHostname(), serviceName),
		)
		if err != nil {
			logger.Log("err", err)
			return err
		} else {
			stdopentracing.SetGlobalTracer(tracer)
		}
	} else {
		stdopentracing.SetGlobalTracer(stdopentracing.NoopTracer{})
	}
	return nil
}

func (z *zipkinInitializer) ReInit(svr Server) error {
	// just do the same init on reinit
	return z.Init(svr)
}
