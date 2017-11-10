package main

import (
	"net/http"

	proto "github.com/carousell/Orion/example/echo/echo_proto"
	"github.com/carousell/Orion/example/echo/service"
	"github.com/carousell/Orion/orion"
	"github.com/gorilla/mux"
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

func encoder(req *http.Request, reqObject interface{}) error {
	vars := mux.Vars(req)
	value, ok := vars["msg"]
	if ok {
		if r, ok := reqObject.(*proto.UpperRequest); ok {
			r.Msg = value
		}
	}
	return nil
}

func main() {
	server := orion.GetDefaultServer("EchoService")
	proto.RegisterEchoServiceOrionServer(&svcFactory{}, server)
	proto.RegisterEchoServiceupperEncoder(server, encoder)
	server.Start()
	server.Wait()
}
