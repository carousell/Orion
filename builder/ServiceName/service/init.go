package service

import (
	"fmt"

	proto "github.com/carousell/Orion/v2/builder/ServiceName/ServiceName_proto"
	"github.com/carousell/Orion/v2/orion"
	"github.com/mitchellh/mapstructure"
)

type svcFactory struct {
}

func (s *svcFactory) NewService(svr orion.Server) interface{} {
	cfg := Config{}
	if c, ok := svr.GetConfig()["echo"]; ok {
		mapstructure.Decode(c, &cfg)
	}
	return GetService(cfg)
}

func (s *svcFactory) DisposeService(svc interface{}) {
	fmt.Println("disposing", svc)
	DestroyService(svc)
}

func GetServiceFactory() orion.ServiceFactory {
	return &svcFactory{}
}

func RegisterOptionals(server orion.Server) {
	proto.RegisterServiceNameUpperEncoder(server, encoder)
	proto.RegisterServiceNameUpperDecoder(server, decoder)
	proto.RegisterServiceNameUpperHandler(server, optionsHandler)
}
