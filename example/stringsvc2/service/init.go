package service

import (
	"strings"

	"github.com/carousell/Orion/orion"
	"github.com/mitchellh/mapstructure"
)

type svcFactory struct{}

func (s *svcFactory) NewService(svr orion.Server) interface{} {
	cfg := Config{}
	if c, ok := svr.GetConfig()[strings.ToLower("config")]; ok {
		mapstructure.Decode(c, &cfg)
	}
	return NewSvc(cfg)
}

func (s *svcFactory) DisposeService(svc interface{}) {
	//do nothing
}

func GetFactory() orion.ServiceFactory {
	return &svcFactory{}
}
