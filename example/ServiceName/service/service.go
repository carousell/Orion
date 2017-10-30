package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/carousell/DataAccessLayer/dal/cassandra"
	"github.com/carousell/DataAccessLayer/dal/es"
	"github.com/carousell/Orion/example/ServiceName/ServiceName_proto"
	"github.com/carousell/Orion/example/ServiceName/service/data"
	"github.com/carousell/Orion/example/ServiceName/service/data/store"
	"github.com/carousell/go-utils/utils/errors"
	"github.com/carousell/go-utils/utils/errors/notifier"
	"google.golang.org/grpc"
)

type sampleServiceImpl struct {
	storageClient data.StorageService
}

func NewService(config Config) SampleService {
	g := sampleServiceImpl{}
	configStr, _ := json.Marshal(config)
	log.Println("Initializing SampleService with", string(configStr))
	g.Init(config)
	return &g
}

func (g *sampleServiceImpl) Init(config Config) {
	c := cassandra.Config{
		Keyspace:                  config.CasKeyspace,
		CassandraHosts:            config.CasHosts,
		CassandraConsistency:      config.CasConsistency,
		CassandraConnectTimeout:   time.Duration(config.CasConnectTimeout) * time.Millisecond,
		CassandraOperationTimeout: time.Duration(config.CasOperationTimeout) * time.Millisecond,
		NumConns:                  config.CasConnections,
	}

	e := es.Config{}
	e.Url = config.ESUrl
	e.Prefix = config.ESPrefix
	e.FakeContext = config.ESFakeContext

	g.storageClient, _ = store.NewClient(c, e)
}

func (g *sampleServiceImpl) Close() {
}

func (g *sampleServiceImpl) Reverse(ctx context.Context, req *ServiceName_proto.EchoRequest) (*ServiceName_proto.EchoResponse, error) {
	resp := new(ServiceName_proto.EchoResponse)
	resp.Msg = "Nope!!"
	return resp, nil
}

func (g *sampleServiceImpl) Echo(ctx context.Context, req *ServiceName_proto.EchoRequest) (*ServiceName_proto.EchoResponse, error) {
	resp := new(ServiceName_proto.EchoResponse)
	resp.Msg = req.Msg
	return resp, nil
}

func (g *sampleServiceImpl) Uppercase(ctx context.Context, req *ServiceName_proto.UppercaseRequest) (*ServiceName_proto.UppercaseResponse, error) {
	resp := new(ServiceName_proto.UppercaseResponse)
	resp.Msg = req.GetMsg()
	resp.Uppercase = strings.ToUpper(req.GetMsg())
	return resp, nil
}

func (g *sampleServiceImpl) AddComment(ctx context.Context, req *ServiceName_proto.AddCommentRequest) (*ServiceName_proto.AddCommentResponse, error) {
	resp, err := g.addComment(ctx, req)
	notifier.Notify(err) // notify errors to rollbar
	return resp, err
}

func (g *sampleServiceImpl) addComment(ctx context.Context, req *ServiceName_proto.AddCommentRequest) (*ServiceName_proto.AddCommentResponse, error) {
	c := strings.TrimSpace(req.GetComment())
	if c == "" {
		return nil, errors.New("comment cant be empty")
	}
	u, err := g.storageClient.AddComment(ctx, c)
	if err != nil {
		return nil, errors.Wrap(err, "storage error")
	}
	resp := new(ServiceName_proto.AddCommentResponse)
	resp.UUID = u
	return resp, nil
}

func (g *sampleServiceImpl) SearchComments(ctx context.Context, req *ServiceName_proto.SearchCommentsRequest) (*ServiceName_proto.SearchCommentsResponse, error) {
	q := strings.TrimSpace(req.GetQuery())
	if q == "" {
		return nil, errors.New("query cant be empty")
	}
	comments, err := g.storageClient.SearchComments(ctx, q)
	if err != nil {
		return nil, errors.Wrap(err, "storage error")
	}
	resp := new(ServiceName_proto.SearchCommentsResponse)
	resp.Comments = make([]*ServiceName_proto.SearchCommentsResponse_Comment, 0, len(comments))
	for _, c := range comments {
		comment := new(ServiceName_proto.SearchCommentsResponse_Comment)
		comment.UUID = c.UUID.String
		comment.Comment = c.Msg.String
		resp.Comments = append(resp.Comments, comment)
	}
	return resp, nil
}

func (g *sampleServiceImpl) GetComment(ctx context.Context, req *ServiceName_proto.GetCommentRequest) (*ServiceName_proto.GetCommentResponse, error) {
	resp, err := g.getComment(ctx, req)
	notifier.Notify(err) // notify errors to rollbar
	return resp, err
}

func (g *sampleServiceImpl) getComment(ctx context.Context, req *ServiceName_proto.GetCommentRequest) (*ServiceName_proto.GetCommentResponse, error) {
	u := strings.TrimSpace(req.GetUUID())
	if u == "" {
		return nil, errors.New("uuid cant be empty")
	}
	c, err := g.storageClient.GetComment(ctx, u)
	if err != nil {
		return nil, errors.Wrap(err, "storage error")
	}
	resp := new(ServiceName_proto.GetCommentResponse)
	resp.UUID = u
	resp.Comment = c
	return resp, nil
}

func (g *sampleServiceImpl) GetInterceptors() []grpc.UnaryServerInterceptor {

	return []grpc.UnaryServerInterceptor{
		func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			fmt.Println(info)
			return handler(ctx, req)
		},
	}
}
