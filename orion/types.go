package orion

import (
	"time"

	"google.golang.org/grpc"
)

const (
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
	RegisterService(sd *grpc.ServiceDesc, ss interface{}) error
	//Wait waits for the Server loop to exit
	Wait() error
	//Stop stops the Server
	Stop(timeout time.Duration) error
}

// ServiceFactory is the interface that need to be implemented by client that provides with a new service object
type ServiceFactory interface {
	// NewService function parses configuration recieved from the orion server
	NewService(server Server) interface{}
}

// HystrixInitializer is the interface that needs to be implemented by client for a custom hystrix initializer
type HystrixInitializer interface {
	InitHystrix()
}

// ZipkinInitializer is the interface that needs to be implemented by client for a custom zipkin initializer
type ZipkinInitializer interface {
	InitZipkin()
}

// NewRelicInitializer is the interface that needs to be implemented by client for a custom newrelic initializer
type NewRelicInitializer interface {
	InitNewRelic()
}
