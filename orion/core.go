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
	newrelic "github.com/newrelic/go-agent"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

type svcInfo struct {
	sd *grpc.ServiceDesc
	sf ServiceFactory
}

type handlerInfo struct {
	handler  handlers.Handler
	listener listenerutils.CustomListener
}

//DefaultServerImpl provides a default implementation of orion.Server this can be embeded in custom orion.Server implementations
type DefaultServerImpl struct {
	config Config
	mu     sync.Mutex
	wg     sync.WaitGroup
	nrApp  newrelic.Application
	inited bool

	services map[string]svcInfo
	handlers []handlerInfo
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
		initInitializers(s)
		s.inited = true
	}
}

func (s *DefaultServerImpl) buildHandlers() []handlerInfo {
	hlrs := []handlerInfo{}
	if !s.config.GRPCOnly {
		httpPort := s.config.HTTPPort
		httpListener, err := listenerutils.NewListener("tcp", ":"+httpPort)
		if err != nil {
			log.Println("error", err)
		}
		log.Println("HTTPListnerPort", httpPort)
		handler := handlers.NewHTTPHandler()
		hlrs = append(hlrs, handlerInfo{
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
		hlrs = append(hlrs, handlerInfo{
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
	c := make(chan os.Signal, 5)
	signal.Notify(c, syscall.SIGHUP)
	for sig := range c {
		if sig == syscall.SIGHUP { // only reload config for sighup
			log.Println("signal", "config reloaded on "+sig.String())
			for _, h := range s.handlers {
				h.listener.StopAccept()
				h.handler.Stop(time.Second * 1)
			}
			log.Println("stop", "stopped all handlers")
			readConfig(s.config.OrionServerName)
			log.Println("reload", "reloading all services")
			for _, info := range s.services {
				s.registerService(info.sd, info.sf, true)
			}
			log.Println("reload", "re initing all handlers")
			for _, h := range s.handlers {
				h.listener = h.listener.GetListener()
				go h.handler.Run(h.listener)
			}
		} else {
			// should not happen!
			for _, h := range s.handlers {
				h.listener.CanClose(true)
				h.handler.Stop(time.Second * 5)
			}
			//logger.Log("signal", "terminating on "+sig.String())
			//errc <- errors.New("terminating on " + sig.String())
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
		s.wg.Add(1)
		go func(s *DefaultServerImpl, h handlerInfo) {
			defer s.wg.Done()
			h.handler.Run(h.listener)
		}(s, h)
	}
	if s.config.HotReload {
		go s.signalWatcher()
	}
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

	for _, h := range s.handlers {
		h.handler.Add(sd, ss)
	}

	s.services[sd.ServiceName] = svcInfo{
		sd: sd,
		sf: sf,
	}
	return nil

}

//Stop stops the server
func (s *DefaultServerImpl) Stop(timeout time.Duration) error {
	var wg sync.WaitGroup
	for _, h := range s.handlers {
		wg.Add(1)
		go func(h handlerInfo, timeout time.Duration) {
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
