package service

import (
	"context"

	"github.com/carousell/Orion/example/ServiceName/ServiceName_proto"
)

type SampleService interface {
	Init(config Config)
	Close()

	Uppercase(ctx context.Context, req *ServiceName_proto.UppercaseRequest) (*ServiceName_proto.UppercaseResponse, error)

	AddComment(ctx context.Context, req *ServiceName_proto.AddCommentRequest) (*ServiceName_proto.AddCommentResponse, error)
	SearchComments(ctx context.Context, req *ServiceName_proto.SearchCommentsRequest) (*ServiceName_proto.SearchCommentsResponse, error)
	GetComment(ctx context.Context, req *ServiceName_proto.GetCommentRequest) (*ServiceName_proto.GetCommentResponse, error)
}

type Config struct {
	CasKeyspace         string
	CasHosts            []string
	CasConsistency      string
	CasConnectTimeout   int
	CasOperationTimeout int
	CasConnections      int
	ESUrl               string
	ESPrefix            string
	ESFakeContext       bool
}
