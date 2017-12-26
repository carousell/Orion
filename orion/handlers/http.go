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

	"github.com/carousell/Orion/utils"
	"github.com/carousell/Orion/utils/headers"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	opentracing "github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	svc         *serviceInfo
	method      GRPCMethodHandler
	encoder     Encoder
	decoder     Decoder
	httpHandler HTTPHandler
	httpMethod  []string
	encoderPath string
}

/*
func (p *pathInfo) Clone() *pathInfo {
	return &pathInfo{
		svc:        p.svc,
		method:     p.method,
		encoder:    p.encoder,
		decoder:    p.decoder,
		httpMethod: p.httpMethod,
	}
}
*/

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

func writeRespWithHeaders(resp http.ResponseWriter, status int, data []byte, headers map[string][]string) {
	if headers != nil {
		for key, values := range headers {
			for _, value := range values {
				resp.Header().Add(key, value)
			}
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
	ctx := utils.StartNRTransaction(url, req.Context(), req, resp)
	defer func(resp http.ResponseWriter, ctx context.Context) {
		// panic handler
		if r := recover(); r != nil {
			writeResp(resp, http.StatusInternalServerError, []byte("Internal Server Error!"))
			log.Println("panic", r)
			log.Print(string(debug.Stack()))
			utils.FinishNRTransaction(ctx, fmt.Errorf("Panic!!"))
		}
	}(resp, ctx)
	req = req.WithContext(ctx)
	err := h.serveHTTP(resp, req, url)
	utils.FinishNRTransaction(req.Context(), err)
}

func prepareContext(req *http.Request, info *pathInfo) context.Context {
	ctx := req.Context()
	//initialize headers
	ctx = headers.AddToRequestHeaders(ctx, "", "")
	ctx = headers.AddToResponseHeaders(ctx, "", "")

	// fetch and populate whitelisted headers
	if len(info.svc.requestHeaders) > 0 {
		for _, hdr := range info.svc.requestHeaders {
			ctx = headers.AddToRequestHeaders(ctx, hdr, req.Header.Get(hdr))
		}
	}

	// translate from http zipkin context to gRPC
	wireContext, err := opentracing.GlobalTracer().Extract(
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(req.Header))
	if err == nil {
		md := metautils.ExtractIncoming(ctx)
		opentracing.GlobalTracer().Inject(wireContext, opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(md))
		ctx = md.ToIncoming(ctx)
	}
	return ctx
}

func grpcErrorToHTTP(err error, defaultStatus int, defaultMessage string) (int, string) {
	code := defaultStatus
	msg := defaultMessage
	if s, ok := status.FromError(err); ok {
		switch s.Code() {
		case codes.NotFound:
			code = http.StatusNotFound
			msg = s.Message()
		case codes.InvalidArgument:
			code = http.StatusBadRequest
			msg = s.Message()
		case codes.Unauthenticated:
			code = http.StatusUnauthorized
			msg = s.Message()
		case codes.PermissionDenied:
			code = http.StatusForbidden
			msg = s.Message()
		}
	}
	return code, msg
}

func (h *httpHandler) serveHTTP(resp http.ResponseWriter, req *http.Request, url string) error {
	info, ok := h.paths[url]
	if ok {
		ctx := prepareContext(req, info)

		// httpHandler allows handling entire http request
		if info.httpHandler != nil {
			req = req.WithContext(ctx)
			if info.httpHandler(resp, req) {
				// short circuit if handler has handled request
				return nil
			}
		}

		// decoder func
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

		// make service call
		protoResponse, err := info.method(info.svc.svc, ctx, dec, info.svc.interceptors)

		if info.decoder != nil {
			//apply decoder if any
			info.decoder(resp, decErr, err, protoResponse)
			if decErr != nil {
				return decErr
			}
			return err
		} else {
			if err != nil {
				if decErr != nil {
					writeResp(resp, http.StatusBadRequest, []byte("Bad Request!"))
					return fmt.Errorf("Bad Request!")
				} else {
					code, msg := grpcErrorToHTTP(err, http.StatusInternalServerError, "Internal Server Error!")
					writeResp(resp, code, []byte(msg))
					return fmt.Errorf(msg)
				}
			} else {
				data, err := h.mar.MarshalToString(protoResponse.(proto.Message))
				if err != nil {
					writeResp(resp, http.StatusInternalServerError, []byte("Internal Server Error!"))
					return fmt.Errorf("Internal Server Error!")
				} else {
					ctx = headers.AddToResponseHeaders(ctx, "Content-Type", "application/json")
					hdr := headers.ResponseHeadersFromContext(ctx)
					responseHeaders := processWhitelist(hdr, append(info.svc.responseHeaders, DefaultHTTPResponseHeaders...))
					writeRespWithHeaders(resp, http.StatusOK, []byte(data), responseHeaders)
					return nil
				}
			}
		}
	} else {
		writeResp(resp, http.StatusNotFound, []byte("Not Found!"))
		return fmt.Errorf("Not Found!")
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
			httpMethod: []string{"POST"},
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

func (h *httpHandler) AddEncoder(serviceName, method string, httpMethod []string, path string, encoder Encoder) {
	if h.paths != nil {
		url := generateURL(serviceName, method)
		if info, ok := h.paths[url]; ok {
			info.encoder = encoder
			info.httpMethod = httpMethod
			info.encoderPath = path
			h.paths[path] = info
		} else {
			fmt.Println("url not found", url, h.paths)
		}
	}
}

func (h *httpHandler) AddHTTPHandler(serviceName string, method string, path string, handler HTTPHandler) {
	if h.paths != nil {
		url := generateURL(serviceName, method)
		if info, ok := h.paths[url]; ok {
			info.httpHandler = handler
		} else {
			fmt.Println("url not found", url, h.paths)
		}
	}
}

func (h *httpHandler) AddDecoder(serviceName, method string, decoder Decoder) {
	if h.paths != nil {
		url := generateURL(serviceName, method)
		if info, ok := h.paths[url]; ok {
			info.decoder = decoder
		} else {
			fmt.Println("url not found", url, h.paths)
		}
	}
}

func (h *httpHandler) Run(httpListener net.Listener) error {
	r := mux.NewRouter()
	fmt.Println("Mapped URLs: ")
	for url, info := range h.paths {
		if strings.TrimSpace(info.encoderPath) != "" && info.encoderPath != url {
			continue
		}
		r.Methods(info.httpMethod...).Path(url).Handler(h.getHTTPHandler(url))
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
