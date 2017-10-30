package orion

import (
	"time"

	"google.golang.org/grpc"
)

const (
	BANNER = `
  ___  ____  ___ ___  _   _
 / _ \|  _ \|_ _/ _ \| \ | |
| | | | |_) || | | | |  \| |
| |_| |  _ < | | |_| | |\  |
 \___/|_| \_\___\___/|_| \_|
                            `
)

// Config is the configuration used by Orion core
type Config struct {
	//OrionServerName is the name of this orion server that is tracked
	OrionServerName string
	// GRPCOnly tells orion not to build HTTP/1.1 server and only initializes gRPC server
	GRPCOnly bool
	//HTTPOnly tells orion not to build gRPC server and only initializes HTTP/1.1 server
	HTTPOnly bool
	// HTTPPort is the port to bind for HTTP requests
	HTTPPort int
	// GRPCPost id the port to bind for gRPC requests
	GRPCPort int
	// ReloadOnConfigChange when set reloads the service when it detects configuration update
	ReloadOnConfigChange bool
}

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
	// Configure function parses configuration recieved from the orion server
	NewService(server Server) interface{}
}
