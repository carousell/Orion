package orion

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"sync"

	"github.com/carousell/Orion/orion/handlers"
	"github.com/carousell/go-utils/utils/listnerutils"
	"google.golang.org/grpc"
)

type DefaultServerImpl struct {
	config      Config
	mu          sync.Mutex
	httpHandler handlers.Handler
	wg          sync.WaitGroup
	grpcServer  *grpc.Server
}

func decoder(in interface{}) error {
	if in == nil {
		return errors.New("No input object!")
	}
	t := reflect.TypeOf(in)
	if t.Kind() != reflect.Struct {
		return errors.New("decoder can only deserialize to structs, can not convert " + t.String() + " of kind " + t.Kind().String())
	}
	return nil
}

func (s *DefaultServerImpl) GetConfig() interface{} {
	return nil
}

func (s *DefaultServerImpl) GetOrionConfig() Config {
	return s.config
}

func (s *DefaultServerImpl) Start() {

	if s.config.HTTPOnly && s.config.GRPCOnly {
		panic("Error: at least one GRPC/HTTP server needs to be initialized")
	}
	fmt.Println(BANNER)
	// start http server
	if !s.config.GRPCOnly {
		if s.httpHandler != nil {
			httpPort := strconv.Itoa(s.config.HTTPPort)
			httpListener, err := listnerutils.NewListener("tcp", ":"+httpPort)
			if err != nil {
				log.Println("error", err)
				return
			}
			fmt.Println("HTTPListnerPort", httpPort)
			s.wg.Add(1)
			go func(s *DefaultServerImpl) {
				defer s.wg.Done()
				s.httpHandler.Run(httpListener, nil)
			}(s)
		}
	}
	if !s.config.HTTPOnly {
		grpcPort := strconv.Itoa(s.config.GRPCPort)
		grpcListener, err := listnerutils.NewListener("tcp", ":"+grpcPort)
		if err != nil {
			log.Println("error", err)
			return
		}
		fmt.Println("gRPCListnerPort", grpcPort)
		s.wg.Add(1)
		go func(s *DefaultServerImpl) {
			defer s.wg.Done()
			s.grpcServer.Serve(grpcListener)
		}(s)
	}
}

func (s *DefaultServerImpl) Wait() error {
	s.wg.Wait()
	return nil
}

func (s *DefaultServerImpl) RegisterService(sd *grpc.ServiceDesc, ss interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.grpcServer == nil {
		s.grpcServer = grpc.NewServer()
	}
	if !s.config.GRPCOnly {
		if s.httpHandler == nil {
			s.httpHandler = handlers.NewHTTPHandler()
		}
		s.httpHandler.Add(sd, ss)
	}
	if !s.config.HTTPOnly {
		s.grpcServer.RegisterService(sd, ss)
	}
	return nil
}

func GetDefaultServer() Server {
	return GetDefaultServerWithConfig(BuildDefaultConfig())
}

func GetDefaultServerWithConfig(config Config) Server {
	return &DefaultServerImpl{config: config}
}
