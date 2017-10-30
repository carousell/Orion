package main

import (
	proto "github.com/carousell/Orion/example/echo/echo_proto"
	"github.com/carousell/Orion/example/echo/service"
	"github.com/carousell/Orion/orion"
)

func main() {
	grpcSrvImpl := service.GetService()
	s := orion.GetDefaultServer()
	proto.RegisterEchoServiceOrionServer(grpcSrvImpl, s)
	s.Start()
	s.Wait()
}
