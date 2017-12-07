package main

import (
	"fmt"
	"log"
	"net/http"

	proto "github.com/carousell/Orion/example/echo/echo_proto"
	"github.com/carousell/Orion/example/echo/service"
	"github.com/carousell/Orion/orion"
	"github.com/gorilla/mux"
	"github.com/mitchellh/mapstructure"
)

type svcFactory struct {
}

func (s *svcFactory) NewService(svr orion.Server) interface{} {
	cfg := service.Config{}
	if c, ok := svr.GetConfig()["echo"]; ok {
		mapstructure.Decode(c, &cfg)
	}
	return service.GetService(cfg)
}

func (s *svcFactory) DisposeService(svc interface{}) {
	fmt.Println("disposing", svc)
}

func encoder(req *http.Request, reqObject interface{}) error {
	vars := mux.Vars(req)
	value, ok := vars["msg"]
	if ok {
		if r, ok := reqObject.(*proto.UpperRequest); ok {
			r.Msg = value
		} else if r, ok := reqObject.(*proto.EchoRequest); ok {
			r.Msg = value
			return nil
		}
		return nil
	}
	return fmt.Errorf("Error: invalid url")
}

func decoder(w http.ResponseWriter, decoderError, endpointError error, respObject interface{}) {
	log.Println("serviceReponse", respObject)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Noo Hello world"))
}

func main() {
	server := orion.GetDefaultServer("EchoService")
	proto.RegisterEchoServiceOrionServer(&svcFactory{}, server)
	proto.RegisterEchoServiceUpperEncoder(server, encoder)
	proto.RegisterEchoServiceUpperDecoder(server, decoder)
	server.Start()
	server.Wait()
}
