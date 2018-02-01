package service

import (
	"context"
	"strings"

	proto "github.com/carousell/Orion/example/stringsvc/stringproto"
)

// svc implements proto.StringServiceServer
type svc struct{}

func (s *svc) Upper(ctx context.Context, req *proto.UpperRequest) (*proto.UpperResponse, error) {
	resp := new(proto.UpperResponse)
	resp.Msg = strings.ToUpper(req.GetMsg())
	return resp, nil
}

func (s *svc) Count(ctx context.Context, req *proto.CountRequest) (*proto.CountResponse, error) {
	resp := new(proto.CountResponse)
	resp.Count = int64(len(req.GetMsg()))
	return resp, nil
}
