package orion

import (
	"time"

	"google.golang.org/grpc"
)

const (
	//ProtoGenVersion1_0 is the version of protoc-gen-orion plugin compatible with current code base
	ProtoGenVersion1_0 = true
	//BANNER is the orion banner text
	BANNER = `
  ___  ____  ___ ___  _   _
 / _ \|  _ \|_ _/ _ \| \ | |
| | | | |_) || | | | |  \| |
| |_| |  _ < | | |_| | |\  |
 \___/|_| \_\___\___/|_| \_|
                            `
)

// Server is the interface that needs to be implemented by any orion server
// 'DefaultServerImpl' should be enough for most users.
type Server interface {
	//Start starts the orion server, this is non blocking call
	Start()
	//RegisterService registers the service to origin server
	RegisterService(sd *grpc.ServiceDesc, sf ServiceFactory) error
	//Wait waits for the Server loop to exit
	Wait() error
	//Stop stops the Server
	Stop(timeout time.Duration) error
	//GetOrionConfig returns current orion config
	GetOrionConfig() Config
	//GetConfig returns current config as parsed from the file/defaults
	GetConfig() map[string]interface{}
	//AddInitializers adds the initializers to orion server
	AddInitializers(ins ...Initializer)
}

//Initializer is the interface needed to be implemented by custom initializers
type Initializer interface {
	Init(svr Server) error
	ReInit(svr Server) error
}

// FactoryParams are the parameters used by the ServiceFactory
type FactoryParams struct {
	// ServiceName contains the proto service name
	ServiceName string
	// Version is a counter that is incremented every time a new service object is requested
	// NOTE: version might rollover in long running services
	Version uint64
}

// ServiceFactory is the interface that needs to be implemented by client that provides a new service object for multiple services
// this allows a single struct to implement multiple services
type ServiceFactory interface {
	// NewService function receives the server object for which service has to be initialized
	NewService(svr Server, params FactoryParams) interface{}
	//DisposeService function disposes the service object
	DisposeService(svc interface{}, params FactoryParams)
}
