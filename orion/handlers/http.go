package handlers

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
)

func NewHTTPHandler() Handler {
	return &httpHandler{}
}

func generateUrl(serviceName, method string) string {
	parts := strings.Split(serviceName, ".")
	if len(parts) > 1 {
		serviceName = strings.ToLower(parts[1])
	}
	method = strings.ToLower(method)
	return "/" + serviceName + "/" + method
}

type serviceInfo struct {
	desc         *grpc.ServiceDesc
	svc          interface{}
	interceptors []grpc.UnaryServerInterceptor
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
	before   []RequestFunc
	after    []ServerResponseFunc
	mar      jsonpb.Marshaler
}

func writeResp(resp http.ResponseWriter, status int, data []byte) {
	resp.WriteHeader(status)
	resp.Write(data)
}

// ChainUnaryServer creates a single interceptor out of a chain of many interceptors.
//
// Execution is done in left-to-right order, including passing of context.
// For example ChainUnaryServer(one, two, three) will execute one before two before three, and three
// will see context changes of one and two.
func chainUnaryServer(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	n := len(interceptors)

	if n > 1 {
		lastI := n - 1
		return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			var (
				chainHandler grpc.UnaryHandler
				curI         int
			)

			chainHandler = func(currentCtx context.Context, currentReq interface{}) (interface{}, error) {
				if curI == lastI {
					return handler(currentCtx, currentReq)
				}
				curI++
				return interceptors[curI](currentCtx, currentReq, info, chainHandler)
			}

			return interceptors[0](ctx, req, info, chainHandler)
		}
	}

	if n == 1 {
		return interceptors[0]
	}

	// n == 0; Dummy interceptor maintained for backward compatibility to avoid returning nil.
	return func(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		return handler(ctx, req)
	}
}

func (h *httpHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	defer func(resp http.ResponseWriter) {
		// panic handler
		if r := recover(); r != nil {
			writeResp(resp, http.StatusInternalServerError, []byte("Internal Server Error!"))
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
		protoResponse, err := info.method(info.svc.svc, ctx, dec, chainUnaryServer(info.svc.interceptors...))
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

	interceptor, ok := ss.(Interceptor)
	if ok {
		svcInfo.interceptors = interceptor.GetInterceptors()
	}

	h.services[sd.ServiceName] = svcInfo

	// TODO recover in case of error
	for _, m := range sd.Methods {
		url := generateUrl(sd.ServiceName, m.MethodName)
		h.paths[url] = pathInfo{
			method: GRPCMethodHandler(m.Handler),
			svc:    svcInfo,
		}
		h.r.Methods("POST").Path(url).Handler(h)
	}
	return nil
}

func (h *httpHandler) Run(httpListener net.Listener, httpSrv *http.Server) error {
	fmt.Println("Mapped URLs: ")
	for url, _ := range h.paths {
		fmt.Println("\tPOST", url)
	}
	if httpSrv == nil {
		httpSrv = &http.Server{
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			Handler:      h.r,
		}
	}
	return httpSrv.Serve(httpListener)
}
