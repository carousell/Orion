// Code generated by protoc-gen-orion. DO NOT EDIT.
// source: standalone_mode/example.proto

package example

import (
	orion "github.com/carousell/Orion/v2/orion"
	example "github.com/carousell/Orion/v2/protoc-gen-orion/testprotos/standalone_mode"
)

// If you see error please update your orion-protoc-gen by running 'go get -u github.com/carousell/Orion/v2/protoc-gen-orion'
var _ = orion.ProtoGenVersion1_0

// Encoders

// Handlers

// Decoders

// Streams

// RegisterExampleServiceOrionServer registers ExampleService to Orion server
// Services need to pass either ServiceFactory or ServiceFactoryV2 implementation
func RegisterExampleServiceOrionServer(sf interface{}, orionServer orion.Server) error {
	err := orionServer.RegisterService(&example.ExampleService_ServiceDesc, sf)
	if err != nil {
		return err
	}

	return nil
}

// DefaultEncoder
func RegisterExampleServiceDefaultEncoder(svr orion.Server, encoder orion.Encoder) {
	orion.RegisterDefaultEncoder(svr, "ExampleService", encoder)
}

// DefaultDecoder
func RegisterExampleServiceDefaultDecoder(svr orion.Server, decoder orion.Decoder) {
	orion.RegisterDefaultDecoder(svr, "ExampleService", decoder)
}

