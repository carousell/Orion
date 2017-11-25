package handlers

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/carousell/Orion/utils/headers"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
)

var (
	DefaultHTTPResponseHeaders = []string{
		"Content-Type",
	}
)

//HTTPHandlerConfig is the configuration for HTTP Handler
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
	serviceName = strings.ToLower(serviceName)
	parts := strings.Split(serviceName, ".")
	if len(parts) > 1 {
		serviceName = parts[1]
	}
	method = strings.ToLower(method)
	return "/" + serviceName + "/" + method
}

func generateProtoURL(serviceName, method string) string {
	return "/" + serviceName + "/" + method
}

type serviceInfo struct {
	desc            *grpc.ServiceDesc
	svc             interface{}
	interceptors    grpc.UnaryServerInterceptor
	requestHeaders  []string
	responseHeaders []string
}

type pathInfo struct {
	svc        *serviceInfo
	method     GRPCMethodHandler
	encoder    Encoder
	httpMethod string
}

func (p *pathInfo) Clone() *pathInfo {
	return &pathInfo{
		svc:        p.svc,
		method:     p.method,
		encoder:    p.encoder,
		httpMethod: p.httpMethod,
	}
}

type httpHandler struct {
	mu       sync.Mutex
	paths    map[string]*pathInfo
	mar      jsonpb.Marshaler
	svr      *http.Server
	protoURL bool
}

func writeResp(resp http.ResponseWriter, status int, data []byte) {
	writeRespWithHeaders(resp, status, data, nil)
}

func writeRespWithHeaders(resp http.ResponseWriter, status int, data []byte, headers map[string]string) {
	if headers != nil {
		for k, v := range headers {
			resp.Header().Add(k, v)
		}
	}
	resp.WriteHeader(status)
	resp.Write(data)
}

func (h *httpHandler) getHTTPHandler(url string) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		h.ServeHTTP(resp, req, url)
	}
}

func (h *httpHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request, url string) {
	defer func(resp http.ResponseWriter) {
		// panic handler
		if r := recover(); r != nil {
			writeResp(resp, http.StatusInternalServerError, []byte("Internal Server Error!"))
			log.Println("panic", r)
			log.Print(string(debug.Stack()))
		}
	}(resp)

	ctx := req.Context()

	info, ok := h.paths[url]
	if ok {
		var decErr error
		dec := func(r interface{}) error {
			if info.encoder == nil {
				protoReq := r.(proto.Message)
				decErr = jsonpb.Unmarshal(req.Body, protoReq)
				return decErr
			}
			decErr = info.encoder(req, r)
			return decErr
		}
		//initialize headers
		ctx = headers.AddToRequestHeaders(ctx, "", "")
		ctx = headers.AddToResponseHeaders(ctx, "", "")
		// fetch and populate whitelisted headers
		if len(info.svc.requestHeaders) > 0 {
			for _, hdr := range info.svc.requestHeaders {
				ctx = headers.AddToRequestHeaders(ctx, hdr, req.Header.Get(hdr))
			}
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
				ctx = headers.AddToResponseHeaders(ctx, "Content-Type", "application/json")
				hdr := headers.ResponseHeadersFromContext(ctx)
				responseHeaders := processWhitelist(hdr.GetAll(), append(info.svc.responseHeaders, DefaultHTTPResponseHeaders...))
				writeRespWithHeaders(resp, http.StatusOK, []byte(data), responseHeaders)
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
		h.paths = make(map[string]*pathInfo)
	}

	svcInfo := &serviceInfo{
		desc: sd,
		svc:  ss,
	}

	svcInfo.interceptors = getInterceptors(ss)
	if headers, ok := ss.(WhitelistedHeaders); ok {
		svcInfo.requestHeaders = headers.GetRequestHeaders()
		svcInfo.responseHeaders = headers.GetResponseHeaders()
	}

	// TODO recover in case of error
	for _, m := range sd.Methods {
		info := &pathInfo{
			method:     GRPCMethodHandler(m.Handler),
			svc:        svcInfo,
			httpMethod: "POST",
		}
		url := generateURL(sd.ServiceName, m.MethodName)
		h.paths[url] = info

		if h.protoURL {
			protoURL := generateProtoURL(sd.ServiceName, m.MethodName)
			h.paths[protoURL] = info
		}
	}
	return nil
}

func (h *httpHandler) AddEncoder(serviceName, method, httpMethod string, path string, encoder Encoder) {
	if h.paths != nil {
		url := generateURL(serviceName, method)
		if info, ok := h.paths[url]; ok {
			i := info.Clone()
			i.encoder = encoder
			i.httpMethod = httpMethod
			delete(h.paths, url)
			h.paths[path] = i
		} else {
			fmt.Println("url not found", url, h.paths)
		}
	}
}

func (h *httpHandler) Run(httpListener net.Listener) error {
	r := mux.NewRouter()
	fmt.Println("Mapped URLs: ")
	for url, info := range h.paths {
		r.Methods(info.httpMethod).Path(url).Handler(h.getHTTPHandler(url))
		fmt.Println("\t", info.httpMethod, url)
	}
	h.svr = &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      r,
	}
	return h.svr.Serve(httpListener)
}

func (h *httpHandler) Stop(timeout time.Duration) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	ctx, can := context.WithTimeout(context.Background(), timeout)
	defer can()
	h.svr.Shutdown(ctx)
	return nil
}
