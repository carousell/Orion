package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	proto "github.com/carousell/Orion/builder/ServiceName/ServiceName_proto"
	"github.com/carousell/Orion/builder/ServiceName/service"
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
	service.DestroyService(svc)
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

func decoder(ctx context.Context, w http.ResponseWriter, decoderError, endpointError error, respObject interface{}) {
	log.Println("serviceReponse", respObject)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Noo Hello world"))
}

func optionsHandler(w http.ResponseWriter, req *http.Request) bool {
	if strings.ToLower(req.Method) == "options" {
		// do something like CORS handling
		w.Header().Set("Test-Header", "testing some data")
		return true
	}
	return false
}

func main() {
	server := orion.GetDefaultServer("EchoService")
	proto.RegisterServiceNameOrionServer(&svcFactory{}, server)
	proto.RegisterServiceNameUpperEncoder(server, encoder)
	proto.RegisterServiceNameUpperDecoder(server, decoder)
	proto.RegisterServiceNameUpperHandler(server, optionsHandler)
	server.Start()
	server.Wait()
}
