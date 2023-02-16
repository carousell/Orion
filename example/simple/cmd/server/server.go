package main

import (
	"fmt"

	"github.com/carousell/Orion/v2/example/simple/service"
	proto "github.com/carousell/Orion/v2/example/simple/simple_proto"
	"github.com/carousell/Orion/v2/orion"
)

type svcFactory struct {
}

func (s *svcFactory) NewService(svr orion.Server) interface{} {
	return service.GetService()
}

func (s *svcFactory) DisposeService(svc interface{}) {
	fmt.Println("disposing", svc)
}

func main() {
	server := orion.GetDefaultServer("SimpleService")
	proto.RegisterSimpleServiceOrionServer(&svcFactory{}, server)
	server.Start()
	server.Wait()
}
