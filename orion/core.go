package orion

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/carousell/Orion/orion/handlers"
	"github.com/carousell/go-utils/utils/listnerutils"
	"google.golang.org/grpc"
)

type DefaultServerImpl struct {
	config      Config
	mu          sync.Mutex
	httpHandler handlers.Handler
	grpcHandler handlers.Handler
	wg          sync.WaitGroup
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
				s.httpHandler.Run(httpListener)
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
			s.grpcHandler.Run(grpcListener)
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
	if !s.config.GRPCOnly {
		if s.httpHandler == nil {
			s.httpHandler = handlers.NewHTTPHandler()
		}
		s.httpHandler.Add(sd, ss)
	}
	if !s.config.HTTPOnly {
		if s.grpcHandler == nil {
			s.grpcHandler = handlers.NewGRPCHandler()
		}
		s.grpcHandler.Add(sd, ss)
	}
	return nil
}

func (s *DefaultServerImpl) Stop(timeout time.Duration) error {
	var wg sync.WaitGroup
	if s.grpcHandler != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.grpcHandler.Stop(timeout)
		}()
	}
	if s.httpHandler != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.grpcHandler.Stop(timeout)
		}()
	}
	wg.Wait()
	return nil
}

func GetDefaultServer() Server {
	return GetDefaultServerWithConfig(BuildDefaultConfig())
}

func GetDefaultServerWithConfig(config Config) Server {
	return &DefaultServerImpl{config: config}
}
