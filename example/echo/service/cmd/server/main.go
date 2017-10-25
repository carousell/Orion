package main

import (
	stdlog "log"
	"os"

	proto "github.com/carousell/Orion/example/echo/echo_proto"
	"github.com/carousell/Orion/example/echo/service"
	"github.com/carousell/Orion/orion"
	"github.com/carousell/go-utils/utils/errors/notifier"
	"github.com/carousell/go-utils/utils/listnerutils"
	"github.com/go-kit/kit/log"
	newrelic "github.com/newrelic/go-agent"
	stdopentracing "github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
)

// these are local vars used through the life span of this service
var (
	logger      log.Logger
	tracer      stdopentracing.Tracer
	newrelicApp newrelic.Application
)

func init() {
	logger = log.NewLogfmtLogger(os.Stdout)
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	logger = log.With(logger, "caller", log.DefaultCaller)
	stdlog.SetOutput(log.NewStdlibAdapter(logger))
}

func main() {
	//grpcPort := viper.GetString(getServiceLower() + ".GRPCPort")
	grpcPort := "8281"
	logger.Log("GRPCaddr", grpcPort)
	grpcListener, err := listnerutils.NewListener("tcp", ":"+grpcPort)
	if err != nil {
		logger.Log("error", err)
		notifier.Notify(err)
		return
	}
	// gRPC transport.
	//glogger := log.With(logger, "transport", "gRPC")
	grpcSrvImpl := service.GetService()
	grpcSrv := grpc.NewServer()
	proto.RegisterEchoServiceServer(grpcSrv, grpcSrvImpl)
	s := orion.GetServer()
	proto.RegisterEchoServiceOrionServer(grpcSrv, grpcSrvImpl, s)
	grpcSrv.Serve(grpcListener)
}
