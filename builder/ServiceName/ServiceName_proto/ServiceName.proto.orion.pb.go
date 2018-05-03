// Code generated by protoc-gen-orion. DO NOT EDIT.
// source: ServiceName.proto

package ServiceName_proto

import (
	orion "github.com/carousell/Orion/orion"
)

// If you see error please update your orion-protoc-gen by running 'go get -u github.com/carousell/Orion/protoc-gen-orion'
var _ = orion.ProtoGenVersion1_0

// Encoders

// RegisterServiceNameUpperEncoder registers the encoder for Upper method in ServiceName
// it registers HTTP  path /api/1.0/upper/{msg} with "GET", "POST", "OPTIONS" methods
func RegisterServiceNameUpperEncoder(svr orion.Server, encoder orion.Encoder) {
	orion.RegisterEncoders(svr, "ServiceName", "Upper", []string{"GET", "POST", "OPTIONS"}, "/api/1.0/upper/{msg}", encoder)
}

// RegisterServiceNameUpperProxyEncoder registers the encoder for UpperProxy method in ServiceName
// it registers HTTP with "POST", "PUT" methods
func RegisterServiceNameUpperProxyEncoder(svr orion.Server, encoder orion.Encoder) {
	orion.RegisterEncoders(svr, "ServiceName", "UpperProxy", []string{"POST", "PUT"}, "", encoder)
}

// Handlers

// RegisterServiceNameUpperHandler registers the handler for Upper method in ServiceName
func RegisterServiceNameUpperHandler(svr orion.Server, handler orion.HTTPHandler) {
	orion.RegisterHandler(svr, "ServiceName", "Upper", "/api/1.0/upper/{msg}", handler)
}

// RegisterServiceNameUpperProxyHandler registers the handler for UpperProxy method in ServiceName
func RegisterServiceNameUpperProxyHandler(svr orion.Server, handler orion.HTTPHandler) {
	orion.RegisterHandler(svr, "ServiceName", "UpperProxy", "", handler)
}

// Decoders

// RegisterServiceNameUpperDecoder registers the decoder for Upper method in ServiceName
func RegisterServiceNameUpperDecoder(svr orion.Server, decoder orion.Decoder) {
	orion.RegisterDecoder(svr, "ServiceName", "Upper", decoder)
}

// RegisterServiceNameUpperProxyDecoder registers the decoder for UpperProxy method in ServiceName
func RegisterServiceNameUpperProxyDecoder(svr orion.Server, decoder orion.Decoder) {
	orion.RegisterDecoder(svr, "ServiceName", "UpperProxy", decoder)
}

// RegisterServiceNameOrionServer registers ServiceName to Orion server
func RegisterServiceNameOrionServer(srv orion.ServiceFactory, orionServer orion.Server) {
	orionServer.RegisterService(&_ServiceName_serviceDesc, srv)

	RegisterServiceNameUpperEncoder(orionServer, nil)
	RegisterServiceNameUpperProxyEncoder(orionServer, nil)
}
