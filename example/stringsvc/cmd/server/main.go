package main

import (
	"github.com/carousell/Orion/example/stringsvc/service"
	proto "github.com/carousell/Orion/example/stringsvc/stringproto"
	"github.com/carousell/Orion/orion"
)

func main() {
	server := orion.GetDefaultServer("StringService")
	proto.RegisterStringServiceOrionServer(service.GetFactory(), server)
	server.Start()
	server.Wait()
}
