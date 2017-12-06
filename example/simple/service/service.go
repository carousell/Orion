package service

import (
	"context"

	proto "github.com/carousell/Orion/example/simple/simple_proto"
)

func GetService() proto.SimpleServiceServer {
	return &svc{}
}

type svc struct {
}

func (s *svc) Echo(ctx context.Context, req *proto.EchoRequest) (*proto.EchoResponse, error) {
	resp := new(proto.EchoResponse)
	resp.Msg = req.GetMsg()
	return resp, nil
}
