// Code generated by protoc-gen-orion. DO NOT EDIT.
// source: example/example.proto

package example

import (
	orion "github.com/carousell/Orion/orion"
)

// If you see error please update your orion-protoc-gen by running 'go get -u github.com/carousell/Orion/protoc-gen-orion'
var _ = orion.ProtoGenVersion1_0

// Encoders

// RegisterExampleServiceHttpGetEncoder registers the encoder for HttpGet method in ExampleService
// it registers HTTP  path /example with "GET" methods
func RegisterExampleServiceHttpGetEncoder(svr orion.Server, encoder orion.Encoder) {
	orion.RegisterEncoders(svr, "ExampleService", "HttpGet", []string{"GET"}, "/example", encoder)
}

// RegisterExampleServiceAuthGetEncoder registers the encoder for AuthGet method in ExampleService
// it registers HTTP  path /auth/example with "GET" methods
func RegisterExampleServiceAuthGetEncoder(svr orion.Server, encoder orion.Encoder) {
	orion.RegisterEncoders(svr, "ExampleService", "AuthGet", []string{"GET"}, "/auth/example", encoder)
}

// RegisterExampleServiceOptionGetEncoder registers the encoder for OptionGet method in ExampleService
// it registers HTTP  path /option/example with "GET" methods
func RegisterExampleServiceOptionGetEncoder(svr orion.Server, encoder orion.Encoder) {
	orion.RegisterEncoders(svr, "ExampleService", "OptionGet", []string{"GET"}, "/option/example", encoder)
}

// RegisterExampleServiceAuthOptionGetEncoder registers the encoder for AuthOptionGet method in ExampleService
// it registers HTTP  path /auth/option/example with "GET" methods
func RegisterExampleServiceAuthOptionGetEncoder(svr orion.Server, encoder orion.Encoder) {
	orion.RegisterEncoders(svr, "ExampleService", "AuthOptionGet", []string{"GET"}, "/auth/option/example", encoder)
}

// Handlers

// RegisterExampleServiceHttpGetHandler registers the handler for HttpGet method in ExampleService
func RegisterExampleServiceHttpGetHandler(svr orion.Server, handler orion.HTTPHandler) {
	orion.RegisterHandler(svr, "ExampleService", "HttpGet", "/example", handler)
}

// RegisterExampleServiceAuthGetHandler registers the handler for AuthGet method in ExampleService
func RegisterExampleServiceAuthGetHandler(svr orion.Server, handler orion.HTTPHandler) {
	orion.RegisterHandler(svr, "ExampleService", "AuthGet", "/auth/example", handler)
}

// RegisterExampleServiceOptionGetHandler registers the handler for OptionGet method in ExampleService
func RegisterExampleServiceOptionGetHandler(svr orion.Server, handler orion.HTTPHandler) {
	orion.RegisterHandler(svr, "ExampleService", "OptionGet", "/option/example", handler)
}

// RegisterExampleServiceAuthOptionGetHandler registers the handler for AuthOptionGet method in ExampleService
func RegisterExampleServiceAuthOptionGetHandler(svr orion.Server, handler orion.HTTPHandler) {
	orion.RegisterHandler(svr, "ExampleService", "AuthOptionGet", "/auth/option/example", handler)
}

// Decoders

// RegisterExampleServiceHttpGetDecoder registers the decoder for HttpGet method in ExampleService
func RegisterExampleServiceHttpGetDecoder(svr orion.Server, decoder orion.Decoder) {
	orion.RegisterDecoder(svr, "ExampleService", "HttpGet", decoder)
}

// RegisterExampleServiceAuthGetDecoder registers the decoder for AuthGet method in ExampleService
func RegisterExampleServiceAuthGetDecoder(svr orion.Server, decoder orion.Decoder) {
	orion.RegisterDecoder(svr, "ExampleService", "AuthGet", decoder)
}

// RegisterExampleServiceOptionGetDecoder registers the decoder for OptionGet method in ExampleService
func RegisterExampleServiceOptionGetDecoder(svr orion.Server, decoder orion.Decoder) {
	orion.RegisterDecoder(svr, "ExampleService", "OptionGet", decoder)
}

// RegisterExampleServiceAuthOptionGetDecoder registers the decoder for AuthOptionGet method in ExampleService
func RegisterExampleServiceAuthOptionGetDecoder(svr orion.Server, decoder orion.Decoder) {
	orion.RegisterDecoder(svr, "ExampleService", "AuthOptionGet", decoder)
}

// Streams

// {ExampleService StreamGet /stream/example true true GET}

// RegisterExampleServiceOrionServer registers ExampleService to Orion server
// Services need to pass either ServiceFactory or ServiceFactoryV2 implementation
func RegisterExampleServiceOrionServer(sf interface{}, orionServer orion.Server) error {
	err := orionServer.RegisterService(&_ExampleService_serviceDesc, sf)
	if err != nil {
		return err
	}

	RegisterExampleServiceHttpGetEncoder(orionServer, nil)
	RegisterExampleServiceAuthGetEncoder(orionServer, nil)
	RegisterExampleServiceOptionGetEncoder(orionServer, nil)
	RegisterExampleServiceAuthOptionGetEncoder(orionServer, nil)
	orion.RegisterMethodOption(orionServer, "ExampleService", "OptionGet", "IGNORE_NR")
	orion.RegisterMethodOption(orionServer, "ExampleService", "AuthOptionGet", "IGNORE_NR")
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

