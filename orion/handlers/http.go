package handlers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
)

type HTTPHandlerConfig struct {
	EnableProtoURL bool
}

//NewHTTPHandler creates a new HTTP handler
func NewHTTPHandler(config HTTPHandlerConfig) Handler {
	return &httpHandler{
		protoURL: config.EnableProtoURL,
	}
}

func generateURL(serviceName, method string) string {
	parts := strings.Split(serviceName, ".")
	if len(parts) > 1 {
		serviceName = strings.ToLower(parts[1])
	}
	method = strings.ToLower(method)
	return "/" + serviceName + "/" + method
}

func generateProtoUrl(serviceName, method string) string {
	return "/" + serviceName + "/" + method
}

type serviceInfo struct {
	desc         *grpc.ServiceDesc
	svc          interface{}
	interceptors grpc.UnaryServerInterceptor
}

type pathInfo struct {
	svc    serviceInfo
	method GRPCMethodHandler
}

type httpHandler struct {
	mu       sync.Mutex
	paths    map[string]pathInfo
	services map[string]serviceInfo
	r        *mux.Router
	mar      jsonpb.Marshaler
	svr      *http.Server
	protoURL bool
}

func writeResp(resp http.ResponseWriter, status int, data []byte) {
	resp.WriteHeader(status)
	resp.Write(data)
}

func (h *httpHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	defer func(resp http.ResponseWriter) {
		// panic handler
		if r := recover(); r != nil {
			writeResp(resp, http.StatusInternalServerError, []byte("Internal Server Error!"))
			log.Println("panic", r)
			log.Print(string(debug.Stack()))
		}
	}(resp)

	ctx := req.Context()
	url := req.URL.String()

	info, ok := h.paths[url]
	if ok {
		var decErr error
		dec := func(r interface{}) error {
			protoReq := r.(proto.Message)
			decErr = jsonpb.Unmarshal(req.Body, protoReq)
			return decErr
		}
		protoResponse, err := info.method(info.svc.svc, ctx, dec, info.svc.interceptors)
		if err != nil {
			if decErr != nil {
				writeResp(resp, http.StatusBadRequest, []byte("Bad Request!"))
			} else {
				writeResp(resp, http.StatusInternalServerError, []byte("Internal Server Error!"))
			}
		} else {
			data, err := h.mar.MarshalToString(protoResponse.(proto.Message))
			if err != nil {
				writeResp(resp, http.StatusInternalServerError, []byte("Internal Server Error!"))
			} else {
				resp.Header().Set("Content-Type", "application/json")
				writeResp(resp, http.StatusOK, []byte(data))
			}
		}
	} else {
		writeResp(resp, http.StatusNotFound, []byte("Not Found!"))
	}
}

func (h *httpHandler) Add(sd *grpc.ServiceDesc, ss interface{}) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.paths == nil {
		h.paths = make(map[string]pathInfo)
	}
	if h.services == nil {
		h.services = make(map[string]serviceInfo)
	}
	if h.r == nil {
		h.r = mux.NewRouter()
	}

	if _, ok := h.services[sd.ServiceName]; ok {
		return errors.New("Service " + sd.ServiceName + " is already initialized")
	}

	svcInfo := serviceInfo{
		desc: sd,
		svc:  ss,
	}

	svcInfo.interceptors = getInterceptors(ss)

	h.services[sd.ServiceName] = svcInfo

	// TODO recover in case of error
	for _, m := range sd.Methods {
		info := pathInfo{
			method: GRPCMethodHandler(m.Handler),
			svc:    svcInfo,
		}
		url := generateURL(sd.ServiceName, m.MethodName)
		h.paths[url] = info
		h.r.Methods("POST").Path(url).Handler(h)

		if h.protoURL {
			protoUrl := generateProtoUrl(sd.ServiceName, m.MethodName)
			h.paths[protoUrl] = info
			h.r.Methods("POST").Path(protoUrl).Handler(h)
		}
	}
	return nil
}

func (h *httpHandler) Run(httpListener net.Listener) error {
	fmt.Println("Mapped URLs: ")
	for url := range h.paths {
		fmt.Println("\tPOST", url)
	}
	h.svr = &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      h.r,
	}
	return h.svr.Serve(httpListener)
}

func (h *httpHandler) Stop(timeout time.Duration) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	ctx, can := context.WithTimeout(context.Background(), timeout)
	defer can()
	h.svr.Shutdown(ctx)
	// reset known services and router
	h.r = nil
	h.services = nil
	return nil
}
