package main

import (
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	stdlog "log"

	proto "github.com/carousell/Orion/example/ServiceName/ServiceName_proto"
	"github.com/carousell/Orion/orion"
	"github.com/carousell/go-utils/utils/errors"
	"github.com/carousell/go-utils/utils/errors/notifier"
	"github.com/carousell/go-utils/utils/listnerutils"
	"github.com/go-kit/kit/log"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func init() {
	logger = log.NewLogfmtLogger(os.Stdout)
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	logger = log.With(logger, "caller", log.DefaultCaller)
	stdlog.SetOutput(log.NewStdlibAdapter(logger))
	setup()
}

func runPprof() {
	go func() {

		pprofport := viper.GetString(getServiceLower() + ".PprofPort")
		logger.Log("PprofPort", pprofport)
		http.ListenAndServe(":"+pprofport, nil)
	}()
}

func main() {
	// read the configuration
	readConfig()

	// setup error logging
	initNotifier()
	defer notifier.Close()

	// hystrix
	initHystrix()

	//pprof
	runPprof()

	//healthcheck
	setupHealthcheck()

	// newrelic
	var err error
	newrelicApp, err = initNewRelic()
	if err != nil {
		logger.Log("err", err)
		notifier.Notify(err)
	}

	httpPort := viper.GetString(getServiceLower() + ".HttpPort")
	logger.Log("HTTPaddr", httpPort)
	httpListener, err := listnerutils.NewListener("tcp", ":"+httpPort)
	if err != nil {
		logger.Log("error", err)
		notifier.Notify(err)
		return
	}

	grpcPort := viper.GetString(getServiceLower() + ".GRPCPort")
	logger.Log("GRPCaddr", grpcPort)
	grpcListener, err := listnerutils.NewListener("tcp", ":"+grpcPort)
	if err != nil {
		logger.Log("error", err)
		notifier.Notify(err)
		return
	}

	errc := make(chan error)
	// start the main app
	go runner(errc, httpListener, grpcListener)

	// SETUP Interrupt handler.
	c := make(chan os.Signal, 5)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	for sig := range c {
		if sig == syscall.SIGHUP { // only reload config for sighup
			logger.Log("signal", "config reloaded on "+sig.String())
			grpcListener.StopAccept()
			grpcListener = grpcListener.GetListner()
			httpListener.StopAccept()
			httpListener = httpListener.GetListner()
			errc <- errors.New("config reloaded on " + sig.String())
			go runner(errc, httpListener, grpcListener)
		} else {
			httpListener.CanClose(true)
			grpcListener.CanClose(true)
			logger.Log("signal", "terminating on "+sig.String())
			errc <- errors.New("terminating on " + sig.String())
			break
		}
	}

	logger.Log("info", "flusing data, waiting for 10 sec")
	newrelicApp.Shutdown(5 * time.Second)
	time.Sleep(7 * time.Second)
}

func runner(errc chan error, httpListener net.Listener, grpcListener net.Listener) {
	logger.Log("msg", "Starting server")
	defer logger.Log("msg", "Shutting down server")

	readConfig()
	//ctx := context.WithValue(context.Background(), "MY-CLIENT-ID", serviceName)

	// zipkin
	tracer = initZipkin()

	service := getService()
	defer service.Close()

	//endpoints := buildEndpoints(service, tracer, newrelicApp)

	srvErr := make(chan error)

	o := orion.GetDefaultServer("BCD")
	grpcSrv := grpc.NewServer()
	proto.RegisterServiceNameServiceOrionServer(grpcSrv, service.(proto.ServiceNameServiceServer), o)
	o.Start()

	/*
		// HTTP transport.
		hlogger := log.With(logger, "transport", "HTTP")
		h := buildHTTPHandler(ctx, endpoints, hlogger, tracer)
		httpSrv := &http.Server{
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			Handler:      h,
		}
		go func() { srvErr <- httpSrv.Serve(httpListener) }()
	*/

	// gRPC transport.
	/*
		glogger := log.With(logger, "transport", "gRPC")
		grpcSrvImpl := buildGRPCServer(ctx, endpoints, glogger, tracer)
		grpc.EnableTracing = true
		grpcSrv := grpc.NewServer()
		registerGRPCServer(grpcSrv, grpcSrvImpl)

		go func() { srvErr <- grpcSrv.Serve(grpcListener) }()
		defer grpcSrv.Stop()
	*/

	// Run!
	select {
	case chanErr := <-errc:
		logger.Log("reload", chanErr)
	case chanErr := <-srvErr:
		logger.Log("exit", chanErr)
	}

	/*
		cont, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		httpSrv.Shutdown(cont)
	*/
	grpcSrv.GracefulStop()
}
