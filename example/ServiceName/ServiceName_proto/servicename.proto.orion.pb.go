// Code generated by protoc-gen-orion. DO NOT EDIT.
// source: ServiceName.proto

package ServiceName_proto

import (
	orion "github.com/carousell/Orion/orion"
	grpc "google.golang.org/grpc"
)

func RegisterServiceNameServiceOrionServer(s *grpc.Server, srv ServiceNameServiceServer, orionServer orion.Server) {
	orionServer.RegisterService(&_ServiceNameService_serviceDesc, srv)
}

