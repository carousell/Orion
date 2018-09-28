package helpers

import (
	"sync"
	"sync/atomic"

	"github.com/carousell/Orion/orion"
)

type svcInfo struct {
	refCount int32
	svc      interface{}
	create   sync.Once
	destroy  sync.Once
}

type svcFactory struct {
	versions sync.Map
	sf       orion.ServiceFactoryV2
}

func (s *svcFactory) NewService(svr orion.Server, params orion.FactoryParams) interface{} {
	// load or store this version
	obj, _ := s.versions.LoadOrStore(params.Version, &svcInfo{})
	info, _ := obj.(*svcInfo)
	// init this obj if not already inited
	info.create.Do(func() {
		info.svc = s.sf.NewService(svr, params)
	})
	// increase object count
	atomic.AddInt32(&info.refCount, 1)
	return info.svc
}

func (s *svcFactory) DisposeService(svc interface{}, params orion.FactoryParams) {
	obj, ok := s.versions.Load(params.Version)
	if !ok {
		// object is already closed
		return
	}
	info, _ := obj.(*svcInfo)
	// decrease object count
	currentCount := atomic.AddInt32(&info.refCount, -1)
	if currentCount < 1 {
		info.destroy.Do(func() {
			s.sf.DisposeService(svc, params)
		})
		s.versions.Delete(params.Version)
	}
}

// NewSingleServiceFactory creates a new service factory that manages single instance from
// the given ServiceFactory across versions (NewService/DisposeService gets called once per version)
func NewSingleServiceFactory(sf interface{}) (orion.ServiceFactoryV2, error) {
	f, err := orion.ToServiceFactoryV2(sf)
	if err != nil {
		return nil, err
	}
	return &svcFactory{
		sf: f,
	}, nil
}
