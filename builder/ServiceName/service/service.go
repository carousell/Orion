// Package service must implement the generated proto's server interface
package service

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	proto "github.com/carousell/Orion/builder/ServiceName/ServiceName_proto"
	"github.com/carousell/Orion/interceptors"
	"github.com/carousell/Orion/utils/headers"
	"github.com/carousell/Orion/utils/worker"
	"github.com/gorilla/mux"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
)

const (
	//address = "192.168.99.100:9281"
	address = "127.0.0.1:9281"
)

type svc struct {
	appendText string
	debug      bool
	client     proto.ServiceNameClient
	worker     worker.Worker
}

func (s *svc) GetRequestHeaders() []string {
	return []string{}
}

func (s *svc) GetResponseHeaders() []string {
	return []string{"Original-Msg"}
}

func GetService(config Config) proto.ServiceNameServer {
	s := new(svc)
	s.appendText = config.AppendText
	s.debug = config.Debug
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(interceptors.DefaultClientInterceptors(address)...)))
	if err != nil {
		log.Fatalln("did not connect: ", err)
	}
	s.client = proto.NewServiceNameClient(conn)
	wConfig := worker.Config{}
	wConfig.LocalMode = true
	/*
		wConfig.RabbitConfig = new(worker.RabbitMQConfig)
		wConfig.RabbitConfig.Host = "192.168.99.100"
		wConfig.RabbitConfig.QueueName = "test"
		wConfig.RabbitConfig.UserName = "guest"
		wConfig.RabbitConfig.Password = "guest"
	*/
	s.worker = worker.NewWorker(wConfig)
	s.worker.RegisterTask("TestWorker", func(ctx context.Context, payload string) error {
		time.Sleep(time.Millisecond * 200)
		log.Println("worker", payload)
		return nil
	})
	s.worker.RunWorker("Worker", 1)
	return s
}

func DestroyService(obj interface{}) {
	if s, ok := obj.(*svc); ok {
		s.worker.CloseWorker()
	}
}

func (s *svc) Echo(ctx context.Context, req *proto.EchoRequest) (*proto.EchoResponse, error) {
	resp := new(proto.EchoResponse)
	resp.Msg = s.appendText + req.GetMsg()

	httpClient := &http.Client{}
	url := "http://127.0.0.1:9282/api/1.0/upper/" + req.GetMsg()
	httpReq, _ := http.NewRequest("GET", url, nil)
	httpReq = httpReq.WithContext(ctx)

	//log.Println(httpReq)
	httpClient.Do(httpReq)

	r := new(proto.UpperRequest)
	r.Msg = "hello"
	s.client.Upper(ctx, r)
	return resp, nil
}

func (s *svc) Upper(ctx context.Context, req *proto.UpperRequest) (*proto.UpperResponse, error) {
	resp := new(proto.UpperResponse)
	resp.Msg = strings.ToUpper(s.appendText + req.GetMsg())
	/*
		hdrs := headers.RequestHeadersFromContext(ctx)
		if hdrs != nil {
			fmt.Println("All request headers", hdrs.GetAll())
		}
	*/
	headers.AddToResponseHeaders(ctx, "original-msg", req.GetMsg())

	/*
		sp, ctx := spanutils.NewDatastoreSpan(ctx, "Wait", "Wa")
		defer sp.End()
		time.Sleep(100 * time.Millisecond)
	*/
	go s.worker.Schedule(ctx, "TestWorker", req.GetMsg())
	return resp, nil
}

func (s *svc) GetInterceptors() []grpc.UnaryServerInterceptor {
	icpt := []grpc.UnaryServerInterceptor{}
	if s.debug {
		icpt = append(icpt, interceptors.DebugLoggingInterceptor())
	}
	return icpt
}

func (s *svc) UpperProxy(ctx context.Context, req *proto.UpperRequest) (*proto.UpperResponse, error) {
	return s.client.Upper(ctx, req)
}

func encoder(req *http.Request, reqObject interface{}) error {
	vars := mux.Vars(req)
	value, ok := vars["msg"]
	if ok {
		if r, ok := reqObject.(*proto.UpperRequest); ok {
			r.Msg = value
		} else if r, ok := reqObject.(*proto.EchoRequest); ok {
			r.Msg = value
			return nil
		}
		return nil
	}
	return fmt.Errorf("Error: invalid url")
}
func decoder(_ context.Context, w http.ResponseWriter, decoderError, endpointError error, respObject interface{}) {
	log.Println("serviceReponse", respObject)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Noo Hello world"))
}

func optionsHandler(w http.ResponseWriter, req *http.Request) bool {
	if strings.ToLower(req.Method) == "options" {
		// do something like CORS handling
		w.Header().Set("Test-Header", "testing some data")
		return true
	}
	return false
}
