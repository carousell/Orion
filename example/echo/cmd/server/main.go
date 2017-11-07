package main

import (
	proto "github.com/carousell/Orion/example/echo/echo_proto"
	"github.com/carousell/Orion/example/echo/service"
	"github.com/carousell/Orion/orion"
	"github.com/mitchellh/mapstructure"
)

type svcFactory struct {
}

func (s *svcFactory) NewService(config map[string]interface{}) interface{} {
	cfg := service.Config{}
	if c, ok := config["echo"]; ok {
		mapstructure.Decode(c, &cfg)
	}
	return service.GetService(cfg)
}

func main() {
	server := orion.GetDefaultServer("EchoService")
	proto.RegisterEchoServiceOrionServer(&svcFactory{}, server)
	server.Start()
	server.Wait()
}
