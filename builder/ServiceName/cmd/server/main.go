package main

import (
	proto "github.com/carousell/Orion/v2/builder/ServiceName/ServiceName_proto"
	"github.com/carousell/Orion/v2/builder/ServiceName/service"
	"github.com/carousell/Orion/v2/orion"
)

func main() {
	server := orion.GetDefaultServer("EchoService")
	proto.RegisterServiceNameOrionServer(service.GetServiceFactory(), server)
	service.RegisterOptionals(server)
	server.Start()
	server.Wait()
}
