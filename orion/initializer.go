package orion

import (
	"errors"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof" // import pprof
	"os"
	"strings"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/carousell/Orion/utils"
	"github.com/carousell/Orion/utils/httptripper"
	logg "github.com/go-kit/kit/log"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	newrelic "github.com/newrelic/go-agent"
	stdopentracing "github.com/opentracing/opentracing-go"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	// NRApp is the key for New Relic app object
	NRApp = "INIT:NR_APP"
)

var (
	//DefaultInitializers are the initializers applied by orion as default
	DefaultInitializers = []Initializer{
		HystrixInitializer(),
		ZipkinInitializer(),
		NewRelicInitializer(),
		PrometheusInitializer(),
		PprofInitializer(),
		HTTPZipkinInitializer(),
	}
)

//HystrixInitializer returns a Initializer implementation for Hystrix
func HystrixInitializer() Initializer {
	return &hystrixInitializer{}
}

//ZipkinInitializer returns a Initializer implementation for Zipkin
func ZipkinInitializer() Initializer {
	return &zipkinInitializer{}
}

//NewRelicInitializer returns a Initializer implementation for NewRelic
func NewRelicInitializer() Initializer {
	return &newRelicInitializer{}
}

//PrometheusInitializer returns a Initializer implementation for Prometheus
func PrometheusInitializer() Initializer {
	return &prometheusInitializer{}
}

//PprofInitializer returns a Initializer implementation for Pprof
func PprofInitializer() Initializer {
	return &pprofInitializer{}
}

//HTTPZipkinInitializer returns an Initializer implementation for httptripper which appends zipkin trace info to all outgoing HTTP requests
func HTTPZipkinInitializer() Initializer {
	return &httpZipkinInitializer{}
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
	}
	log.Println("NR", "initialized with "+serviceName)
	svr.Store(NRApp, app)
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
		}
		stdopentracing.SetGlobalTracer(tracer)
	} else {
		stdopentracing.SetGlobalTracer(stdopentracing.NoopTracer{})
	}
	return nil
}

func (z *zipkinInitializer) ReInit(svr Server) error {
	// just do the same init on reinit
	return z.Init(svr)
}

type prometheusInitializer struct {
}

func (p *prometheusInitializer) Init(svr Server) error {
	if svr.GetOrionConfig().EnablePrometheus {
		if svr.GetOrionConfig().EnablePrometheusHistogram {
			grpc_prometheus.EnableHandlingTimeHistogram()
		}
		// Register Prometheus metrics handler.
		http.Handle("/metrics", promhttp.Handler())
	}
	return nil
}

func (p *prometheusInitializer) ReInit(svr Server) error {
	return nil
}

type pprofInitializer struct {
}

func (p *pprofInitializer) Init(svr Server) error {

	go func(svr Server) {
		pprofport := svr.GetOrionConfig().PProfport
		log.Println("PprofPort", pprofport)
		http.ListenAndServe(":"+pprofport, nil)
	}(svr)
	return nil
}

func (p *pprofInitializer) ReInit(svr Server) error {
	return nil
}

type httpZipkinInitializer struct {
}

func (h *httpZipkinInitializer) Init(svr Server) error {
	tripper := httptripper.WrapTripper(http.DefaultTransport)
	http.DefaultTransport = tripper
	return nil
}

func (h *httpZipkinInitializer) ReInit(svr Server) error {
	// Do nothing
	return nil
}
