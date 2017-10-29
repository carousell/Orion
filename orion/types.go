package orion

import (
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
	// GRPCOnly tells origin not to build HTTP/1.1 server and only initializes gRPC server
	GRPCOnly bool
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
	// GetConfig fetches application config
	GetConfig() interface{}
	//Start function start the orion server, this will block forever
	Start()
	//RegisterService registers the service to origin server
	RegisterService(sd *grpc.ServiceDesc, ss interface{}) error
}

// OrionServviceInitializer is the interface that need to be implemented by any service that needs
// to pass configuration before initializing a orion service
type OrionServiceInitializer interface {
	// Configure function parses configuration recieved from the orion server
	Init(interface{}) bool
}
