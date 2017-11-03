package main

import (
	proto "github.com/carousell/Orion/example/echo/echo_proto"
	"github.com/carousell/Orion/example/echo/service"
	"github.com/carousell/Orion/orion"
)

type svcFactory struct {
}

func (s *svcFactory) NewService(config map[string]interface{}) interface{} {
	return service.GetService()
}

func main() {
	s := orion.GetDefaultServer("EchoService")
	proto.RegisterEchoServiceOrionServer(&svcFactory{}, s)
	s.Start()
	s.Wait()
}
