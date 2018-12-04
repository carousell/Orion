package orion

import (
	"context"
	"errors"
	"net"
	"net/http"
	_ "net/http/pprof" // import pprof
	"os"
	"strings"
	"time"

	"github.com/afex/hystrix-go/hystrix"
	metricCollector "github.com/afex/hystrix-go/hystrix/metric_collector"
	"github.com/afex/hystrix-go/plugins"
	"github.com/carousell/Orion/utils"
	"github.com/carousell/Orion/utils/errors/notifier"
	"github.com/carousell/Orion/utils/log"
	logg "github.com/go-kit/kit/log"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	newrelic "github.com/newrelic/go-agent"
	stdopentracing "github.com/opentracing/opentracing-go"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	//DefaultInitializers are the initializers applied by orion as default
	DefaultInitializers = []Initializer{
		HystrixInitializer(),
		ZipkinInitializer(),
		NewRelicInitializer(),
		PrometheusInitializer(),
		PprofInitializer(),
		ErrorLoggingInitializer(),
	}
)

//HystrixInitializer returns a Initializer implementation for Hystrix
func HystrixInitializer() Initializer {
	return &hystrixInitializer{}
}

//ErrorLoggingInitializer returns a Initializer implementation for error notifier
func ErrorLoggingInitializer() Initializer {
	return &errorLoggingInitializer{}
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

type hystrixInitializer struct {
}

func (h *hystrixInitializer) Init(svr Server) error {
	config := svr.GetOrionConfig()
	hystrix.DefaultTimeout = 1000 // one sec
	hystrix.DefaultMaxConcurrent = 300
	hystrix.DefaultErrorPercentThreshold = 75
	hystrix.DefaultSleepWindow = 1000
	hystrix.DefaultVolumeThreshold = 75

	if strings.TrimSpace(config.HystrixConfig.StatsdAddr) != "" {
		name := config.OrionServerName + ".hystrix"
		name = strings.Replace(name, "-", "_", 10)

		c, err := plugins.InitializeStatsdCollector(&plugins.StatsdCollectorConfig{
			StatsdAddr: config.HystrixConfig.StatsdAddr,
			Prefix:     name,
		})
		if err == nil {
			metricCollector.Registry.Register(c.NewStatsdCollector)
			log.Info(context.Background(), "HystrixStatsd", config.HystrixConfig.StatsdAddr)
		} else {
			log.Info(context.Background(), "Hystrix", err.Error())
		}

	}
	hystrixStreamHandler := hystrix.NewStreamHandler()
	hystrixStreamHandler.Start()
	port := config.HystrixConfig.Port
	log.Info(context.Background(), "HystrixPort", port)
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
	config := svr.GetOrionConfig().NewRelicConfig
	serviceName := config.ServiceName
	if strings.TrimSpace(serviceName) == "" {
		serviceName = svr.GetOrionConfig().OrionServerName
	}
	nrConfig := newrelic.NewConfig(serviceName, apiKey)
	if config.ExcludeAttributes != nil && len(config.ExcludeAttributes) > 0 {
		nrConfig.Attributes.Enabled = true
		nrConfig.Attributes.Exclude = config.ExcludeAttributes
	}
	if config.IncludeAttributes != nil && len(config.IncludeAttributes) > 0 {
		nrConfig.Attributes.Enabled = true
		nrConfig.Attributes.Include = config.IncludeAttributes
	}
	app, err := newrelic.NewApplication(nrConfig)
	if err != nil {
		log.Error(context.Background(), "nr-error", err)
		return err
	}
	utils.NewRelicApp = app
	log.Info(context.Background(), "NR", "initialized with "+serviceName)
	return nil
}

func (n *newRelicInitializer) ReInit(svr Server) error {
	return n.Init(svr)
}

type zipkinInitializer struct {
	tracer    stdopentracing.Tracer
	collector zipkin.Collector
}

func (z *zipkinInitializer) Init(svr Server) error {

	oldCollector := z.collector

	zipkinAddr := svr.GetOrionConfig().ZipkinConfig.Addr
	serviceName := svr.GetOrionConfig().OrionServerName
	if zipkinAddr != "" {
		logger := logg.NewJSONLogger(os.Stdout)
		logger = logg.With(logger, "time", logg.DefaultTimestampUTC)
		logger.Log("zipkin-addr", zipkinAddr)
		var err error
		if strings.HasPrefix(zipkinAddr, "http") {
			z.collector, err = zipkin.NewHTTPCollector(
				zipkinAddr,
				zipkin.HTTPLogger(logger),
			)
			zipkin.HTTPBatchSize(1)
		} else {
			z.collector, err = zipkin.NewKafkaCollector(
				strings.Split(zipkinAddr, ","),
				zipkin.KafkaLogger(logger),
			)
		}
		if err != nil {
			logger.Log("err", err)
			return err
		}

		z.tracer, err = zipkin.NewTracer(
			zipkin.NewRecorder(z.collector, true, utils.GetHostname(), serviceName),
		)
		if err != nil {
			logger.Log("err", err)
			return err
		}
		stdopentracing.SetGlobalTracer(z.tracer)
		// close old collector
		if oldCollector != nil {
			go func(oldCollector zipkin.Collector) {
				// close old collector after 5 seconds
				time.Sleep(time.Second * 5)
				oldCollector.Close()
			}(oldCollector)
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
		log.Info(context.Background(), "PprofPort", pprofport)
		http.ListenAndServe(":"+pprofport, nil)
	}(svr)
	return nil
}

func (p *pprofInitializer) ReInit(svr Server) error {
	return nil
}

type errorLoggingInitializer struct{}

func (e *errorLoggingInitializer) Init(svr Server) error {
	env := svr.GetOrionConfig().Env
	// environment for error notification
	notifier.SetEnvironemnt(env)

	// rollbar
	rToken := svr.GetOrionConfig().RollbarToken
	if strings.TrimSpace(rToken) != "" {
		notifier.InitRollbar(rToken, env)
		log.Debug(context.Background(), "rollbarToken", rToken, "env", env)
	}

	//sentry
	sToken := svr.GetOrionConfig().SentryDSN
	if strings.TrimSpace(sToken) != "" {
		notifier.InitSentry(sToken)
		log.Debug(context.Background(), "sentryDSN", rToken, "env", env)
	}
	return nil
}

func (e *errorLoggingInitializer) ReInit(svr Server) error {
	return e.Init(svr)
}
