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
	grpcHandler "github.com/carousell/Orion/orion/handlers/grpc"
	"github.com/carousell/Orion/orion/handlers/http"
	"github.com/carousell/Orion/utils/errors/notifier"
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
	httpMethod  []string
	path        string
	encoder     Encoder
	handler     HTTPHandler
}

type decoderInfo struct {
	serviceName string
	method      string
	decoder     handlers.Decoder
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
	config Config
	mu     sync.Mutex
	wg     sync.WaitGroup
	inited bool

	services     map[string]*svcInfo
	encoders     map[string]*encoderInfo
	decoders     map[string]*decoderInfo
	defDecoders  map[string]handlers.Decoder
	defEncoders  map[string]handlers.Encoder
	options      map[string]*optionInfo
	middlewares  map[string]*middlewareInfo
	handlers     []*handlerInfo
	initializers []Initializer

	dataBag map[string]interface{}
}

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

//AddEncoder is the implementation of handlers.Encodable
func (d *DefaultServerImpl) AddEncoder(serviceName, method string, httpMethod []string, path string, encoder handlers.Encoder) {
	if d.encoders == nil {
		d.encoders = make(map[string]*encoderInfo)
	}
	key := getSvcKey(serviceName, method)
	ei := new(encoderInfo)
	if info, ok := d.encoders[key]; ok {
		ei = info
	} else {
		d.encoders[key] = ei
	}
	ei.serviceName = serviceName
	ei.method = method
	ei.httpMethod = httpMethod
	ei.path = path
	ei.encoder = encoder
}

//AddDefaultEncoder is the implementation of handlers.Encodable
func (d *DefaultServerImpl) AddDefaultEncoder(serviceName string, encoder Encoder) {
	if d.defEncoders == nil {
		d.defEncoders = make(map[string]handlers.Encoder)
	}
	d.defEncoders[serviceName] = encoder
}

//AddHTTPHandler is the implementation of handlers.HTTPInterceptor
func (d *DefaultServerImpl) AddHTTPHandler(serviceName string, method string, path string, handler handlers.HTTPHandler) {
	if d.encoders == nil {
		d.encoders = make(map[string]*encoderInfo)
	}
	key := getSvcKey(serviceName, method)
	ei := new(encoderInfo)
	if info, ok := d.encoders[key]; ok {
		ei = info
	} else {
		d.encoders[key] = ei
	}
	ei.serviceName = serviceName
	ei.method = method
	ei.path = path
	ei.handler = handler
}

//AddDecoder is the implementation of handlers.Decodable
func (d *DefaultServerImpl) AddDecoder(serviceName, method string, decoder handlers.Decoder) {
	if d.decoders == nil {
		d.decoders = make(map[string]*decoderInfo)
	}
	d.decoders[serviceName+":"+method] = &decoderInfo{
		serviceName: serviceName,
		method:      method,
		decoder:     decoder,
	}
}

//AddDefaultDecoder is the implementation of handlers.Decodable
func (d *DefaultServerImpl) AddDefaultDecoder(serviceName string, decoder Decoder) {
	if d.defDecoders == nil {
		d.defDecoders = make(map[string]handlers.Decoder)
	}
	d.defDecoders[serviceName] = decoder
}

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

func (d *DefaultServerImpl) init(reload bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.inited != true {
		d.initHandlers()
		d.initInitializers(reload)
		d.inited = true
	}
}

func (d *DefaultServerImpl) initInitializers(reload bool) {

	if d.initializers == nil {
		d.initializers = DefaultInitializers
	}
	d.processInitializers(reload)
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

func (d *DefaultServerImpl) buildHandlers() []*handlerInfo {
	hlrs := []*handlerInfo{}
	if !d.config.GRPCOnly {
		httpPort := d.config.HTTPPort
		httpListener, err := listenerutils.NewListener("tcp", ":"+httpPort)
		if err != nil {
			log.Println("error", err)
		}
		log.Println("HTTPListnerPort", httpPort)
		config := http.HandlerConfig{
			EnableProtoURL: d.config.EnableProtoURL,
		}
		handler := http.NewHTTPHandler(config)
		hlrs = append(hlrs, &handlerInfo{
			handler:  handler,
			listener: httpListener,
		})
	}
	if !d.config.HTTPOnly {
		grpcPort := d.config.GRPCPort
		grpcListener, err := listenerutils.NewListener("tcp", ":"+grpcPort)
		if err != nil {
			log.Println("error", err)
		}
		log.Println("gRPCListnerPort", grpcPort)
		handler := grpcHandler.NewGRPCHandler(grpcHandler.GRPCConfig{})
		hlrs = append(hlrs, &handlerInfo{
			handler:  handler,
			listener: grpcListener,
		})
	}
	return hlrs
}

func (d *DefaultServerImpl) initHandlers() {
	d.handlers = d.buildHandlers()
}

func (d *DefaultServerImpl) signalWatcher() {
	// Setup interrupt handler.
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP)
	for sig := range c {
		if sig == syscall.SIGHUP { // only reload config for sighup
			log.Println("signal", "config reloaded on "+sig.String())
			// relaod config
			err := readConfig(d.config.OrionServerName)
			if err != nil {
				notifier.NotifyWithLevel(err, "critical", "Error parsing config not reloading services")
				log.Println("Error", err, "msg", "not reloading services")
				continue
			}

			// reload initializers
			d.processInitializers(true)

			// reload services
			oldServices := []*svcInfo{}
			for _, info := range d.services {
				d.registerService(info.sd, info.sf, true)
				oldServices = append(oldServices, info)
			}

			// reload handlers
			for _, h := range d.handlers {
				d.startHandler(h, true)
			}

			//dispose the older service object
			for _, info := range oldServices {
				info.sf.DisposeService(info.ss)
			}
		} else {
			// should not happen!
			for _, h := range d.handlers {
				h.listener.CanClose(true)
				h.handler.Stop(time.Second * 5)
			}
			break
		}
		log.Println("signal", "all actions complete")
	}
}

//Start starts the orion server
func (d *DefaultServerImpl) Start() {
	fmt.Println(BANNER)
	if d.config.HTTPOnly && d.config.GRPCOnly {
		panic("Error: at least one GRPC or HTTP server needs to be initialized")
	}

	for _, h := range d.handlers {
		d.startHandler(h, false)
	}
	if d.config.HotReload {
		go d.signalWatcher()
	}
}

func (d *DefaultServerImpl) startHandler(h *handlerInfo, reload bool) {
	if reload {
		h.listener.StopAccept()
		h.handler.Stop(time.Second * 1)
		h.listener = h.listener.GetListener()
	}

	//Add all services first
	for _, info := range d.services {
		h.handler.Add(info.sd, info.ss)
	}

	//Add all encoders
	if e, ok := h.handler.(handlers.Encodeable); ok {
		for _, ei := range d.encoders {
			e.AddEncoder(ei.serviceName, ei.method, ei.httpMethod, ei.path, ei.encoder)
			if ei.handler != nil {
				if i, ok := h.handler.(handlers.HTTPInterceptor); ok {
					i.AddHTTPHandler(ei.serviceName, ei.method, ei.path, ei.handler)
				}
			}
		}
	}

	//Add all options
	if e, ok := h.handler.(handlers.Optionable); ok {
		for _, oi := range d.options {
			e.AddOption(oi.serviceName, oi.method, oi.option)
		}
	}

	//Add all default encoders
	if e, ok := h.handler.(handlers.Encodeable); ok {
		for svc, enc := range d.defEncoders {
			e.AddDefaultEncoder(svc, enc)
		}
	}

	//Add all decoders
	if e, ok := h.handler.(handlers.Decodable); ok {
		for _, di := range d.decoders {
			e.AddDecoder(di.serviceName, di.method, di.decoder)
		}
	}

	//Add all default decoders
	if e, ok := h.handler.(handlers.Decodable); ok {
		for svc, dec := range d.defDecoders {
			e.AddDefaultDecoder(svc, dec)
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
//Note: this is only callled from code generated by orion plugin
func (d *DefaultServerImpl) RegisterService(sd *grpc.ServiceDesc, sf ServiceFactory) error {
	d.init(false) // make sure its called before lock
	return d.registerService(sd, sf, false)
}

func (d *DefaultServerImpl) registerService(sd *grpc.ServiceDesc, sf ServiceFactory, reload bool) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.services == nil {
		d.services = make(map[string]*svcInfo)
	}

	_, ok := d.services[sd.ServiceName]
	if ok && !reload {
		return errors.New("error: service " + sd.ServiceName + " already added!")
	}
	// create a obejct from factory and check types
	ss := sf.NewService(d)
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
	return &DefaultServerImpl{
		config: config,
	}
}
