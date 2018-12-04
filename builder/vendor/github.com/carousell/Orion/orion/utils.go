package orion

// ToServiceFactoryV2 converts ServiceFactory to ServiceFactoryV2
func ToServiceFactoryV2(sf interface{}) (ServiceFactoryV2, error) {
	if sf == nil {
		return nil, ErrNil
	}
	// first check for ServiceFactoryV2
	if f, ok := sf.(ServiceFactoryV2); ok {
		return f, nil
	} else if f, ok := sf.(ServiceFactory); ok {
		return &sfv2{f}, nil
	}
	return nil, ErrNotServiceFactory
}

type sfv2 struct {
	sf ServiceFactory
}

func (s *sfv2) NewService(svr Server, params FactoryParams) interface{} {
	return s.sf.NewService(svr)
}

func (s *sfv2) DisposeService(svc interface{}, params FactoryParams) {
	s.sf.DisposeService(svc)
}
