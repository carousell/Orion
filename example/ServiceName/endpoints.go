package ServiceName

import (
	"context"

	"github.com/carousell/Orion/example/ServiceName/ServiceName_proto"
	"github.com/carousell/Orion/example/ServiceName/service"
	"github.com/carousell/go-utils/utils"
	"github.com/carousell/go-utils/utils/errors"
	"github.com/carousell/go-utils/utils/errors/notifier"
	"github.com/carousell/healthcheck"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/tracing/opentracing"
	newrelic "github.com/newrelic/go-agent"
	stdopentracing "github.com/opentracing/opentracing-go"
)

type Endpoints struct {
	// always keep the healthcheck endpoint
	HealthCheck endpoint.Endpoint

	// Echo endpoint
	Echo endpoint.Endpoint
	//Uppercse endpoint
	Uppercase endpoint.Endpoint

	// AddComment endpoint
	AddComment endpoint.Endpoint
	// Search comments endpoint
	SearchComments endpoint.Endpoint
	//GetComment endpoint
	GetComment endpoint.Endpoint
}

func MakeUppercaseEndpoint(svc service.SampleService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(*ServiceName_proto.UppercaseRequest)
		if !ok {
			return nil, errors.New("cant cast request to *ServiceName_proto.UppercaseRequest")
		}
		v, err := svc.Uppercase(ctx, req)
		if err != nil {
			return v, err
		}
		return v, nil
	}
}

func MakeAddCommentEndpoint(svc service.SampleService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(*ServiceName_proto.AddCommentRequest)
		if !ok {
			return nil, errors.New("cant cast request to *ServiceName_proto.AddCommentRequest")
		}
		v, err := svc.AddComment(ctx, req)
		if err != nil {
			return v, err
		}
		return v, nil
	}
}

func MakeSearchCommentsEndpoint(svc service.SampleService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(*ServiceName_proto.SearchCommentsRequest)
		if !ok {
			return nil, errors.New("cant cast request to *ServiceName_proto.SearchCommentsRequest")
		}
		v, err := svc.SearchComments(ctx, req)
		if err != nil {
			return v, err
		}
		return v, nil
	}
}

func MakeGetCommentEndpoint(svc service.SampleService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(*ServiceName_proto.GetCommentRequest)
		if !ok {
			return nil, errors.New("cant cast request to *ServiceName_proto.GetCommentRequest")
		}
		v, err := svc.GetComment(ctx, req)
		if err != nil {
			return v, err
		}
		return v, nil
	}
}

func MakeHelathCheckEndpoint() endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		// TODO add check for switching state
		h := healthcheck.GetHealthCheck()
		if !h.IsHealthy() {
			return nil, errors.New("Node Unhealthy")
		}
		return "OK", nil
	}
}

func BuildEndpoints(service service.SampleService, tracer stdopentracing.Tracer, newrelicApp newrelic.Application) Endpoints {
	return Endpoints{
		// get calls
		Uppercase: applyMiddleware(MakeUppercaseEndpoint(service), "Uppercase", tracer, newrelicApp),

		SearchComments: applyMiddleware(MakeSearchCommentsEndpoint(service), "SearchComments", tracer, newrelicApp),
		AddComment:     applyMiddleware(MakeAddCommentEndpoint(service), "AddComment", tracer, newrelicApp),
		GetComment:     applyMiddleware(MakeGetCommentEndpoint(service), "GetComment", tracer, newrelicApp),

		// healthcheck
		HealthCheck: applyMiddleware(MakeHelathCheckEndpoint(), "HealthCheck", tracer, newrelicApp),
	}
}

func applyMiddleware(endpoint endpoint.Endpoint, name string, tracer stdopentracing.Tracer, app newrelic.Application) endpoint.Endpoint {
	return notifier.TracingMiddleware()(notifier.NotifyingMiddleware(name)(utils.NewRelicMiddleware(name, app)(opentracing.TraceServer(tracer, name)(endpoint))))
}
