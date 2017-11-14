package orion

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"reflect"
	"sync"
	"syscall"
	"time"

	"github.com/carousell/Orion/orion/handlers"
	"github.com/carousell/Orion/utils/listenerutils"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

type svcInfo struct {
	sd *grpc.ServiceDesc
	sf ServiceFactory
	ss interface{}
}

type handlerInfo struct {
	handler  handlers.Handler
	listener listenerutils.CustomListener
}

type encoderInfo struct {
	serviceName string
	method      string
	httpMethod  string
	path        string
	encoder     handlers.Encoder
}

//DefaultServerImpl provides a default implementation of orion.Server this can be embeded in custom orion.Server implementations
type DefaultServerImpl struct {
	config Config
	mu     sync.Mutex
	wg     sync.WaitGroup
	inited bool

	services     map[string]svcInfo
	encoders     map[string]encoderInfo
	handlers     []*handlerInfo
	initializers []Initializer

	dataBag map[string]interface{}
}

//Store stores values for use by initializers
func (d *DefaultServerImpl) Store(key string, value interface{}) {
	if d.dataBag == nil {
		d.dataBag = make(map[string]interface{})
	}
	if value == nil {
		if _, found := d.dataBag[key]; found {
			delete(d.dataBag, key)
		}
	}
}

//Fetch fetches values for use by initializers
func (d *DefaultServerImpl) Fetch(key string) (value interface{}, found bool) {
	value, found = d.dataBag[key]
	return
}

func (d *DefaultServerImpl) AddEncoder(serviceName, method, httpMethod string, path string, encoder handlers.Encoder) {
	if d.encoders == nil {
		d.encoders = make(map[string]encoderInfo)
	}
	d.encoders[path] = encoderInfo{
		serviceName: serviceName,
		method:      method,
		httpMethod:  httpMethod,
		path:        path,
		encoder:     encoder,
	}
}

//GetOrionConfig returns current orion config
//NOTE: this config can not be modifies
func (s *DefaultServerImpl) GetOrionConfig() Config {
	return s.config
}

func (s *DefaultServerImpl) init(reload bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.inited != true {
		s.initHandlers()
		s.initInitializers(reload)
		s.inited = true
	}
}

func (d *DefaultServerImpl) initInitializers(reload bool) {

	var in interface{}
	in = d

	// pre init
	if i, ok := in.(PreInitializer); ok {
		i.PreInit()
	}

	if d.initializers == nil {
		d.initializers = DefaultInitializers()
	}
	d.processInitializers(reload)

	// post init
	if i, ok := in.(PostInitializer); ok {
		i.PostInit()
	}
}

func (d *DefaultServerImpl) processInitializers(reload bool) {
	if d.initializers == nil {
		return
	}
	for _, in := range d.initializers {
		if in != nil {
			if reload {
				in.ReInit(d)
			} else {
				in.Init(d)
			}
		}
	}
}

func (s *DefaultServerImpl) buildHandlers() []*handlerInfo {
	hlrs := []*handlerInfo{}
	if !s.config.GRPCOnly {
		httpPort := s.config.HTTPPort
		httpListener, err := listenerutils.NewListener("tcp", ":"+httpPort)
		if err != nil {
			log.Println("error", err)
		}
		log.Println("HTTPListnerPort", httpPort)
		config := handlers.HTTPHandlerConfig{
			EnableProtoURL: s.config.EnableProtoURL,
		}
		handler := handlers.NewHTTPHandler(config)
		hlrs = append(hlrs, &handlerInfo{
			handler:  handler,
			listener: httpListener,
		})
	}
	if !s.config.HTTPOnly {
		grpcPort := s.config.GRPCPort
		grpcListener, err := listenerutils.NewListener("tcp", ":"+grpcPort)
		if err != nil {
			log.Println("error", err)
		}
		log.Println("gRPCListnerPort", grpcPort)
		handler := handlers.NewGRPCHandler()
		hlrs = append(hlrs, &handlerInfo{
			handler:  handler,
			listener: grpcListener,
		})
	}
	return hlrs
}

func (s *DefaultServerImpl) initHandlers() {
	s.handlers = s.buildHandlers()
}

func (s *DefaultServerImpl) signalWatcher() {
	// SETUP Interrupt handler.
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP)
	for sig := range c {
		if sig == syscall.SIGHUP { // only reload config for sighup
			log.Println("signal", "config reloaded on "+sig.String())
			readConfig(s.config.OrionServerName)
			for _, info := range s.services {
				s.registerService(info.sd, info.sf, true)
			}
			for _, h := range s.handlers {
				s.startHandler(h, true)
			}
		} else {
			// should not happen!
			for _, h := range s.handlers {
				h.listener.CanClose(true)
				h.handler.Stop(time.Second * 5)
			}
			break
		}
		log.Println("signal", "all actions complete")
	}
}

//Start starts the orion server
func (s *DefaultServerImpl) Start() {
	fmt.Println(BANNER)
	if s.config.HTTPOnly && s.config.GRPCOnly {
		panic("Error: at least one GRPC or HTTP server needs to be initialized")
	}

	for _, h := range s.handlers {
		s.startHandler(h, false)
	}
	if s.config.HotReload {
		go s.signalWatcher()
	}
}

func (s *DefaultServerImpl) startHandler(h *handlerInfo, reload bool) {
	if reload {
		h.listener.StopAccept()
		h.handler.Stop(time.Second * 1)
		h.listener = h.listener.GetListener()
	}
	for _, info := range s.services {
		h.handler.Add(info.sd, info.ss)
	}

	for _, ei := range s.encoders {
		if e, ok := h.handler.(handlers.Encodeable); ok {
			e.AddEncoder(ei.serviceName, ei.method, ei.httpMethod, ei.path, ei.encoder)
		}
	}
	s.wg.Add(1)
	go func(s *DefaultServerImpl, h *handlerInfo) {
		defer s.wg.Done()
		err := h.handler.Run(h.listener)
		log.Println("exited", h, err)
	}(s, h)
}

// Wait waits for all the serving servers to quit
func (s *DefaultServerImpl) Wait() error {
	s.wg.Wait()
	return nil
}

//RegisterService registers a service from a generated proto file
//Note: this is only callled from code generated by orion plugin
func (s *DefaultServerImpl) RegisterService(sd *grpc.ServiceDesc, sf ServiceFactory) error {
	s.init(false) // make sure its called before lock
	return s.registerService(sd, sf, false)
}

func (s *DefaultServerImpl) registerService(sd *grpc.ServiceDesc, sf ServiceFactory, reload bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.services == nil {
		s.services = make(map[string]svcInfo)
	}

	_, ok := s.services[sd.ServiceName]
	if ok && !reload {
		return errors.New("error: service " + sd.ServiceName + " already added!")
	}
	// create a obejct from factory and check types
	ss := sf.NewService(viper.AllSettings())
	ht := reflect.TypeOf(sd.HandlerType).Elem()
	st := reflect.TypeOf(ss)
	if !st.Implements(ht) {
		return fmt.Errorf("Orion.Server.RegisterService found the handler of type %v that does not satisfy %v", st, ht)
	}

	s.services[sd.ServiceName] = svcInfo{
		sd: sd,
		sf: sf,
		ss: ss,
	}
	return nil

}

//AddInitializers adds the initializers to orion server
func (d *DefaultServerImpl) AddInitializers(ins ...Initializer) {
	if d.initializers == nil {
		d.initializers = make([]Initializer, 0)
	}
	if len(ins) > 0 {
		d.initializers = append(d.initializers, ins...)
	}

}

//GetConfig returns current config as parsed from the file/defaults
func (d *DefaultServerImpl) GetConfig() map[string]interface{} {
	return viper.AllSettings()
}

//Stop stops the server
func (s *DefaultServerImpl) Stop(timeout time.Duration) error {
	var wg sync.WaitGroup
	for _, h := range s.handlers {
		wg.Add(1)
		go func(h *handlerInfo, timeout time.Duration) {
			defer wg.Done()
			h.handler.Stop(timeout)
		}(h, timeout)
	}
	wg.Wait()
	return nil
}

//GetDefaultServer returns a default server object that can be directly used to start orion server
func GetDefaultServer(name string) Server {
	return GetDefaultServerWithConfig(BuildDefaultConfig(name))
}

//GetDefaultServerWithConfig returns a default server object that uses provided configuration
func GetDefaultServerWithConfig(config Config) Server {
	return &DefaultServerImpl{config: config}
}
