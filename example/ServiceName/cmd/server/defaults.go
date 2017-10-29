package main

import (
	"context"
	"net/http"

	"google.golang.org/grpc"

	svc "github.com/carousell/Orion/example/ServiceName"
	proto "github.com/carousell/Orion/example/ServiceName/ServiceName_proto"
	mainservice "github.com/carousell/Orion/example/ServiceName/service"
	"github.com/go-kit/kit/log"
	newrelic "github.com/newrelic/go-agent"
	stdopentracing "github.com/opentracing/opentracing-go"
	"github.com/spf13/viper"
)

const (
	serviceName string = "ServiceName"
)

func setConfigDefaults() {
	viper.SetDefault(getServiceLower()+".GRPCPort", "9281")
	viper.SetDefault(getServiceLower()+".HttpPort", "9282")
	viper.SetDefault(getServiceLower()+".HystrixPort", "9283")
	viper.SetDefault(getServiceLower()+".PprofPort", "9284")
	viper.SetDefault(getServiceLower()+".ZipkinAddr", "http://10.200.0.7:9411/api/v1/spans")
	viper.SetDefault(getServiceLower()+".env", "dev")
	viper.SetDefault("rollbar.token", "")
	// all custom defaults go here
	setCustomConfigDefaults()
}

func setCustomConfigDefaults() {
	mainservice.SetSvcDefaults()
}

func buildEndpoints(service mainservice.SampleService, tracer stdopentracing.Tracer, newrelicApp newrelic.Application) svc.Endpoints {
	return svc.BuildEndpoints(service, tracer, newrelicApp)
}

func buildHTTPHandler(ctx context.Context, endpoints svc.Endpoints, logger log.Logger, tracer stdopentracing.Tracer) http.Handler {
	return svc.MakeHTTPHandler(ctx, endpoints, logger, tracer)
}

func buildGRPCServer(ctx context.Context, endpoints svc.Endpoints, logger log.Logger, tracer stdopentracing.Tracer) proto.ServiceNameServiceServer {
	return svc.MakeGRPCServer(ctx, endpoints, logger, tracer)
}

func registerGRPCServer(s *grpc.Server, srv proto.ServiceNameServiceServer) {
	proto.RegisterServiceNameServiceServer(s, srv)
}

func getService() mainservice.SampleService {
	return mainservice.NewService(buildSvcConfig())
}

func buildSvcConfig() mainservice.Config {
	return mainservice.BuildSvcConfig()
}
