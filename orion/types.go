package orion

import (
	"time"

	"github.com/carousell/Orion/orion/handlers"
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

// ServiceFactory is the interface that need to be implemented by client that provides with a new service object
type ServiceFactory interface {
	// NewService function receives the server obejct for which service has to be initialized
	NewService(Server) interface{}
	//DisposeService function disposes the service object
	DisposeService(svc interface{})
}

//Encoder is the function type needed for request encoders
type Encoder = handlers.Encoder

//Decoder is the function type needed for request decoders
type Decoder = handlers.Decoder

//HTTPHandler is the http interceptor
type HTTPHandler = handlers.HTTPHandler
