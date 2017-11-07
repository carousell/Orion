package main

import (
	proto "github.com/carousell/Orion/example/echo/echo_proto"
	"github.com/carousell/Orion/example/echo/service"
	"github.com/carousell/Orion/orion"
	"github.com/spf13/viper"
)

type svcFactory struct {
}

func (s *svcFactory) NewService(config map[string]interface{}) interface{} {
	c := service.Config{
		AppendText: viper.GetString("echo.AppendText"),
	}
	return service.GetService(c)
}

func main() {
	s := orion.GetDefaultServer("EchoService")
	proto.RegisterEchoServiceOrionServer(&svcFactory{}, s)
	s.Start()
	s.Wait()
}
