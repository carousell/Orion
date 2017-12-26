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

//DefaultServerImpl provides a default implementation of orion.Server this can be embeded in custom orion.Server implementations
type DefaultServerImpl struct {
	config Config
	mu     sync.Mutex
	wg     sync.WaitGroup
	inited bool

	services     map[string]*svcInfo
	encoders     map[string]*encoderInfo
	decoders     map[string]*decoderInfo
	handlers     []*handlerInfo
	initializers []Initializer

	dataBag map[string]interface{}
}

//AddEncoder is the implementation of handlers.Encodable
func (d *DefaultServerImpl) AddEncoder(serviceName, method string, httpMethod []string, path string, encoder handlers.Encoder) {
	if d.encoders == nil {
		d.encoders = make(map[string]*encoderInfo)
	}
	ei := new(encoderInfo)
	if info, ok := d.encoders[path]; ok {
		ei = info
	} else {
		d.encoders[path] = ei
	}
	ei.serviceName = serviceName
	ei.method = method
	ei.httpMethod = httpMethod
	ei.path = path
	ei.encoder = encoder
}

//AddHTTPHandler is the implementation of handlers.HTTPInterceptor
func (d *DefaultServerImpl) AddHTTPHandler(serviceName string, method string, path string, handler handlers.HTTPHandler) {
	if d.encoders == nil {
		d.encoders = make(map[string]*encoderInfo)
	}
	ei := new(encoderInfo)
	if info, ok := d.encoders[path]; ok {
		ei = info
	} else {
		d.encoders[path] = ei
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
		config := handlers.HTTPHandlerConfig{
			EnableProtoURL: d.config.EnableProtoURL,
		}
		handler := handlers.NewHTTPHandler(config)
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
		handler := handlers.NewGRPCHandler()
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
	for _, info := range d.services {
		h.handler.Add(info.sd, info.ss)
	}

	for _, ei := range d.encoders {
		if e, ok := h.handler.(handlers.Encodeable); ok {
			e.AddEncoder(ei.serviceName, ei.method, ei.httpMethod, ei.path, ei.encoder)
			if ei.handler != nil {
				if i, ok := h.handler.(handlers.HTTPInterceptor); ok {
					i.AddHTTPHandler(ei.serviceName, ei.method, ei.path, ei.handler)
				}
			}
		}
	}
	for _, di := range d.decoders {
		if e, ok := h.handler.(handlers.Decodable); ok {
			e.AddDecoder(di.serviceName, di.method, di.decoder)
		}
	}
	d.wg.Add(1)
	go func(s *DefaultServerImpl, h *handlerInfo) {
		defer s.wg.Done()
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
	return &DefaultServerImpl{config: config}
}
