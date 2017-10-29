package ServiceName

import (
	"context"
	"errors"

	"github.com/carousell/Orion/example/ServiceName/ServiceName_proto"
	"github.com/carousell/go-utils/utils"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/tracing/opentracing"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	stdopentracing "github.com/opentracing/opentracing-go"
	netcontext "golang.org/x/net/context"
)

type grpcServer struct {
	echo      grpctransport.Handler
	uppercase grpctransport.Handler

	searchComments grpctransport.Handler
	addComment     grpctransport.Handler
	getComment     grpctransport.Handler
}

func (g *grpcServer) Echo(c netcontext.Context, req *ServiceName_proto.EchoRequest) (*ServiceName_proto.EchoResponse, error) {
	_, resp, err := g.echo.ServeGRPC(c, req)
	if err != nil {
		return nil, err
	}
	return resp.(*ServiceName_proto.EchoResponse), nil
}

func (g *grpcServer) Reverse(c netcontext.Context, req *ServiceName_proto.EchoRequest) (*ServiceName_proto.EchoResponse, error) {
	resp := new(ServiceName_proto.EchoResponse)
	for _, v := range req.Msg {
		resp.Msg = string(v) + resp.Msg
	}
	return resp, nil
}

func (g *grpcServer) Uppercase(c netcontext.Context, req *ServiceName_proto.UppercaseRequest) (*ServiceName_proto.UppercaseResponse, error) {
	_, resp, err := g.uppercase.ServeGRPC(c, req)
	if err != nil {
		return nil, err
	}
	return resp.(*ServiceName_proto.UppercaseResponse), nil
}

func (g *grpcServer) SearchComments(c netcontext.Context, req *ServiceName_proto.SearchCommentsRequest) (*ServiceName_proto.SearchCommentsResponse, error) {
	_, resp, err := g.searchComments.ServeGRPC(c, req)
	if err != nil {
		return nil, err
	}
	return resp.(*ServiceName_proto.SearchCommentsResponse), nil
}

func (g *grpcServer) AddComment(c netcontext.Context, req *ServiceName_proto.AddCommentRequest) (*ServiceName_proto.AddCommentResponse, error) {
	_, resp, err := g.addComment.ServeGRPC(c, req)
	if err != nil {
		return nil, err
	}
	return resp.(*ServiceName_proto.AddCommentResponse), nil
}

func (g *grpcServer) GetComment(c netcontext.Context, req *ServiceName_proto.GetCommentRequest) (*ServiceName_proto.GetCommentResponse, error) {
	_, resp, err := g.getComment.ServeGRPC(c, req)
	if err != nil {
		return nil, err
	}
	return resp.(*ServiceName_proto.GetCommentResponse), nil
}

func DecodeGRPCRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	return grpcReq, nil
}

func EncodeGRPCResponse(_ context.Context, grpcResp interface{}) (interface{}, error) {
	return grpcResp, nil
}

func DecodeEchoGRPCRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	return grpcReq, nil
}

func EncodeEchoGRPCResponse(_ context.Context, grpcResp interface{}) (interface{}, error) {
	resp := new(ServiceName_proto.EchoResponse)
	if s, ok := grpcResp.(string); ok {
		resp.Msg = s
		return resp, nil
	}
	return nil, errors.New("invalid EncodeEchoGRPCResponse")
}

func newGRPCServer(name string, logger log.Logger, tracer stdopentracing.Tracer, e endpoint.Endpoint, dec grpctransport.DecodeRequestFunc, enc grpctransport.EncodeResponseFunc) *grpctransport.Server {
	logger = log.With(logger, "Endpoint", name)
	options := []grpctransport.ServerOption{
		grpctransport.ServerErrorLogger(logger),
		grpctransport.ServerBefore(opentracing.GRPCToContext(tracer, name, logger)),
	}
	return grpctransport.NewServer(
		utils.EndpointLoggingMiddleware(logger)(e),
		dec,
		enc,
		options...,
	)
}

func MakeGRPCServer(ctx context.Context, endpoints Endpoints, logger log.Logger, tracer stdopentracing.Tracer) ServiceName_proto.ServiceNameServiceServer {
	return &grpcServer{
		echo: newGRPCServer(
			"Echo",
			logger,
			tracer,
			endpoints.Echo,
			DecodeEchoGRPCRequest,
			EncodeEchoGRPCResponse,
		),
		uppercase: newGRPCServer(
			"Uppercase",
			logger,
			tracer,
			endpoints.Uppercase,
			DecodeGRPCRequest,
			EncodeGRPCResponse,
		),
		searchComments: newGRPCServer(
			"SearchComments",
			logger,
			tracer,
			endpoints.SearchComments,
			DecodeGRPCRequest,
			EncodeGRPCResponse,
		),
		addComment: newGRPCServer(
			"AddComment",
			logger,
			tracer,
			endpoints.AddComment,
			DecodeGRPCRequest,
			EncodeGRPCResponse,
		),
		getComment: newGRPCServer(
			"GetComment",
			logger,
			tracer,
			endpoints.GetComment,
			DecodeGRPCRequest,
			EncodeGRPCResponse,
		),
	}
}
