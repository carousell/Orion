package orion

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/viper"
	"google.golang.org/grpc"

	"github.com/carousell/Orion/v2/orion/handlers"
	grpcHandler "github.com/carousell/Orion/v2/orion/handlers/grpc"
	"github.com/carousell/Orion/v2/utils/listenerutils"
	"github.com/carousell/Orion/v2/utils/log"
)

var (
	//ErrNil when the passed argument is nil
	ErrNil = errors.New("nil argument passed")
)

type svcInfo struct {
	sd *grpc.ServiceDesc
	sf ServiceFactoryV2
	ss interface{}
}

type handlerInfo struct {
	handler  handlers.Handler
	listener listenerutils.CustomListener
}

type optionInfo struct {
	serviceName string
	method      string
	option      string
}

type middlewareInfo struct {
	serviceName string
	method      string
	middlewares []string
}

//DefaultServerImpl provides a default implementation of orion.Server this can be embedded in custom orion.Server implementations
type DefaultServerImpl struct {
	config                    Config
	mu                        sync.Mutex
	wg                        sync.WaitGroup
	grpcUnknownServiceHandler grpc.StreamHandler
	inited                    bool

	services     map[string]*svcInfo
	options      map[string]*optionInfo
	middlewares  map[string]*middlewareInfo
	handlers     []*handlerInfo
	initializers []Initializer
	version      uint64
}

//AddMiddleware adds middlewares for particular service/method
func (d *DefaultServerImpl) AddMiddleware(serviceName string, method string, middlewares ...string) {
	if d.middlewares == nil {
		d.middlewares = make(map[string]*middlewareInfo)
	}
	if len(middlewares) > 0 {
		key := getSvcKey(serviceName, method)
		if info, ok := d.middlewares[key]; ok {
			if info.middlewares != nil {
				info.middlewares = append(info.middlewares, middlewares...)
			} else {
				info.middlewares = middlewares
			}
		} else {
			mi := new(middlewareInfo)
			mi.serviceName = serviceName
			mi.method = method
			mi.middlewares = middlewares
			d.middlewares[key] = mi
		}
	}
}

func getSvcKey(serviceName, method string) string {
	return serviceName + "-" + method
}

//AddOption adds a option for the particular service/method
func (d *DefaultServerImpl) AddOption(serviceName, method, option string) {
	if d.options == nil {
		d.options = make(map[string]*optionInfo)
	}
	d.options[serviceName+":"+method] = &optionInfo{
		serviceName: serviceName,
		method:      method,
		option:      option,
	}
}

//GetOrionConfig returns current orion config
//NOTE: this config can not be modifies
func (d *DefaultServerImpl) GetOrionConfig() Config {
	return d.config
}

func (d *DefaultServerImpl) init() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.inited != true {
		d.initHandlers()
		d.initInitializers()
		d.inited = true
	}
}

func (d *DefaultServerImpl) initInitializers() {
	if d.initializers == nil {
		return
	}
	for _, in := range d.initializers {
		if in != nil {
			in.Init(d)
		}
	}
}

func (d *DefaultServerImpl) buildHandlers() []*handlerInfo {
	hlrs := []*handlerInfo{}
	grpcPort := d.config.GRPCPort
	grpcListener, err := listenerutils.NewListener("tcp", ":"+grpcPort)
	if err != nil {
		log.Info(context.Background(), "grpcListener", "could not create listener", "error", err)
	}
	log.Info(context.Background(), "gRPCListenerPort", grpcPort)
	config := grpcHandler.Config{
		CommonConfig: handlers.CommonConfig{
			DisableDefaultInterceptors: d.config.DisableDefaultInterceptors,
		},
		UnknownServiceHandler: d.grpcUnknownServiceHandler,
		MaxRecvMsgSize:        d.config.MaxRecvMsgSize,
	}
	handler := grpcHandler.NewGRPCHandler(config)
	hlrs = append(hlrs, &handlerInfo{
		handler:  handler,
		listener: grpcListener,
	})
	return hlrs
}

func (d *DefaultServerImpl) initHandlers() {
	d.handlers = d.buildHandlers()
}

func (d *DefaultServerImpl) signalWatcher() {
	// Setup interrupt handler.
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	for sig := range c {
		if sig == syscall.SIGTERM || sig == syscall.SIGINT {
			log.Info(context.Background(), "signal", "starting shutdown on "+sig.String())
			d.Stop(30 * time.Second)
			break
		} else {
			// should not happen!
			for _, h := range d.handlers {
				h.listener.CanClose(true)
				h.handler.Stop(time.Second * 5)
			}
			break
		}
		log.Info(context.Background(), "signal", "all actions complete")
	}
}

//Start starts the orion server
func (d *DefaultServerImpl) Start() {
	fmt.Println(BANNER)

	for _, h := range d.handlers {
		d.startHandler(h)
	}
	go d.signalWatcher()
}

func (d *DefaultServerImpl) startHandler(h *handlerInfo) {
	//Add all services first
	for _, info := range d.services {
		h.handler.Add(info.sd, info.ss)
	}

	//Add all options
	if e, ok := h.handler.(handlers.Optionable); ok {
		for _, oi := range d.options {
			e.AddOption(oi.serviceName, oi.method, oi.option)
		}
	}

	// Add all middlewares
	if e, ok := h.handler.(handlers.Middlewareable); ok {
		for _, mi := range d.middlewares {
			e.AddMiddleware(mi.serviceName, mi.method, mi.middlewares...)
		}
	}

	d.wg.Add(1)
	go func(d *DefaultServerImpl, h *handlerInfo) {
		defer d.wg.Done()
		h.handler.Run(h.listener)
	}(d, h)
}

// Wait waits for all the serving servers to quit
func (d *DefaultServerImpl) Wait() error {
	d.wg.Wait()
	return nil
}

//RegisterService registers a service from a generated proto file
//Note: this is only called from code generated by orion plugin
func (d *DefaultServerImpl) RegisterService(sd *grpc.ServiceDesc, sf ServiceFactoryV2) error {
	d.init() // make sure its called before lock

	d.mu.Lock()
	defer d.mu.Unlock()

	if d.services == nil {
		d.services = make(map[string]*svcInfo)
	}

	_, ok := d.services[sd.ServiceName]
	if ok {
		return errors.New("error: service " + sd.ServiceName + " already added!")
	}

	params := FactoryParams{
		ServiceName: sd.ServiceName,
		Version:     d.version,
	}
	// create an object from factory and check types
	ss := sf.NewService(d, params)
	ht := reflect.TypeOf(sd.HandlerType).Elem()
	st := reflect.TypeOf(ss)
	if !st.Implements(ht) {
		return fmt.Errorf("Orion.Server.RegisterService found the handler of type %v that does not satisfy %v", st, ht)
	}

	d.services[sd.ServiceName] = &svcInfo{
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
func (d *DefaultServerImpl) Stop(timeout time.Duration) error {
	var wg sync.WaitGroup
	for _, h := range d.handlers {
		h.listener.CanClose(true)
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
func GetDefaultServer(name string, opts ...DefaultServerOption) Server {
	server := &DefaultServerImpl{
		config: BuildDefaultConfig(name),
	}
	for _, opt := range opts {
		opt.apply(server)
	}
	return server
}

//GetDefaultServerWithConfig returns a default server object that uses provided configuration
func GetDefaultServerWithConfig(config Config) Server {
	return &DefaultServerImpl{
		config: config,
	}
}

type DefaultServerOption interface {
	apply(*DefaultServerImpl)
}

// WithGrpcUnknownHandler returns a DefaultServerOption which sets
// UnknownServiceHandler option in grpc server
func WithGrpcUnknownHandler(grpcUnknownServiceHandler grpc.StreamHandler) DefaultServerOption {
	return newFuncDefaultServerOption(func(h *DefaultServerImpl) {
		h.grpcUnknownServiceHandler = grpcUnknownServiceHandler
	})
}

type funcDefaultServerOption struct {
	f func(options *DefaultServerImpl)
}

func (fdso *funcDefaultServerOption) apply(ds *DefaultServerImpl) {
	fdso.f(ds)
}

func newFuncDefaultServerOption(f func(options *DefaultServerImpl)) *funcDefaultServerOption {
	return &funcDefaultServerOption{
		f: f,
	}
}
