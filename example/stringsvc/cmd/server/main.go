package main

import (
	"github.com/carousell/Orion/v2/example/stringsvc/service"
	proto "github.com/carousell/Orion/v2/example/stringsvc/stringproto"
	"github.com/carousell/Orion/v2/orion"
)

func main() {
	server := orion.GetDefaultServer("StringService")
	proto.RegisterStringServiceOrionServer(service.GetFactory(), server)
	server.Start()
	server.Wait()
}
