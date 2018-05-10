package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/carousell/Orion/orion/modifiers"
	"github.com/carousell/Orion/utils"
	"github.com/carousell/Orion/utils/errors/notifier"
	"github.com/carousell/Orion/utils/headers"
	"github.com/carousell/Orion/utils/options"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"github.com/mitchellh/mapstructure"
	opentracing "github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	// DefaultHTTPResponseHeaders are response headers that are whitelisted by default
	DefaultHTTPResponseHeaders = []string{
		"Content-Type",
	}
)

//HTTPHandlerConfig is the configuration for HTTP Handler
type HTTPHandlerConfig struct {
	CommonConfig
	EnableProtoURL bool
}

//NewHTTPHandler creates a new HTTP handler
func NewHTTPHandler(config HTTPHandlerConfig) Handler {
	return &httpHandler{
		config: config,
	}
}

func cleanSvcName(serviceName string) string {
	serviceName = strings.ToLower(serviceName)
	parts := strings.Split(serviceName, ".")
	if len(parts) > 1 {
		serviceName = parts[1]
	}
	return serviceName
}

func generateURL(serviceName, method string) string {
	serviceName = cleanSvcName(serviceName)
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

type httpHandler struct {
	mu          sync.Mutex
	paths       map[string]*pathInfo
	defEncoders map[string]Encoder
	defDecoders map[string]Decoder
	mar         jsonpb.Marshaler
	svr         *http.Server
	config      HTTPHandlerConfig
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
	var err error
	ctx := utils.StartNRTransaction(url, req.Context(), req, resp)
	defer func(resp http.ResponseWriter, ctx context.Context, t time.Time) {
		// panic handler
		if r := recover(); r != nil {
			writeResp(resp, http.StatusInternalServerError, []byte("Internal Server Error!"))
			log.Println("panic", r, "path", req.URL.String(), "method", req.Method, "took", time.Since(t))
			log.Print(string(debug.Stack()))
			err := fmt.Errorf("panic: %s", r)
			utils.FinishNRTransaction(ctx, err)
			notifier.NotifyWithLevel(err, "critical", req.URL.String(), ctx)
		} else {
			log.Println("path", req.URL.String(), "method", req.Method, "error", err, "took", time.Since(t))
		}
	}(resp, ctx, time.Now())
	req = req.WithContext(ctx)
	ctx, err = h.serveHTTP(resp, req, url)
	if modifiers.HasDontLogError(ctx) {
		utils.FinishNRTransaction(req.Context(), nil)
	} else {
		notifier.Notify(err, req.URL.String(), ctx)
		utils.FinishNRTransaction(req.Context(), err)
	}
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

	if values, found := req.Header["Content-Type"]; found {
		for _, value := range values {
			headers.AddToRequestHeaders(ctx, "Content-Type", value)
		}
	}
	if values, found := req.Header["Accept"]; found {
		for _, value := range values {
			headers.AddToRequestHeaders(ctx, "Accept", value)
		}
	}

	// populate options
	ctx = options.AddToOptions(ctx, modifiers.RequestHTTP, true)

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

// GrpcErrorToHTTP converts gRPC error code into HTTP response status code.
// See: https://github.com/googleapis/googleapis/blob/master/google/rpc/code.proto
func GrpcErrorToHTTP(err error, defaultStatus int, defaultMessage string) (int, string) {
	code := defaultStatus
	msg := defaultMessage
	if s, ok := status.FromError(err); ok {
		msg = s.Message()
		switch s.Code() {
		case codes.NotFound:
			code = http.StatusNotFound
		case codes.InvalidArgument:
			code = http.StatusBadRequest
		case codes.Unauthenticated:
			code = http.StatusUnauthorized
		case codes.PermissionDenied:
			code = http.StatusForbidden
		case codes.OK:
			code = http.StatusOK
		case codes.Canceled:
			code = http.StatusRequestTimeout
		case codes.Unknown:
			code = http.StatusInternalServerError
		case codes.DeadlineExceeded:
			code = http.StatusGatewayTimeout
		case codes.AlreadyExists:
			code = http.StatusConflict
		case codes.ResourceExhausted:
			code = http.StatusTooManyRequests
		case codes.FailedPrecondition:
			code = http.StatusBadRequest
		case codes.Aborted:
			code = http.StatusConflict
		case codes.OutOfRange:
			code = http.StatusBadRequest
		case codes.Unimplemented:
			code = http.StatusNotImplemented
		case codes.Internal:
			code = http.StatusInternalServerError
		case codes.Unavailable:
			code = http.StatusServiceUnavailable
		case codes.DataLoss:
			code = http.StatusInternalServerError
		}
	}
	return code, msg
}

// DefaultEncoder encodes a HTTP request if none are registered. This encoder
// populates the proto message with URL route variables or fields from a JSON
// body if either are available.
func DefaultEncoder(req *http.Request, r interface{}) error {
	protoReq := r.(proto.Message)
	// check and map url params to request
	params := mux.Vars(req)
	if len(params) > 0 {
		mapstructure.Decode(params, protoReq)
	}
	err := jsonpb.Unmarshal(req.Body, protoReq)
	if req.Method == http.MethodGet && err == io.EOF {
		return nil
	}
	return err
}

func (h *httpHandler) serveHTTP(resp http.ResponseWriter, req *http.Request, url string) (context.Context, error) {
	info, ok := h.paths[url]
	if ok {
		ctx := prepareContext(req, info)

		// httpHandler allows handling entire http request
		if info.httpHandler != nil {
			req = req.WithContext(ctx)
			if info.httpHandler(resp, req) {
				// short circuit if handler has handled request
				return ctx, nil
			}
		}

		// decoder func
		var encErr error
		dec := func(r interface{}) error {
			if info.encoder != nil {
				encErr = info.encoder(req, r)
			} else if enc, ok := h.defEncoders[cleanSvcName(info.svc.desc.ServiceName)]; ok {
				// check for default encoder and invoke it
				encErr = enc(req, r)
			} else {
				encErr = DefaultEncoder(req, r)
			}
			return encErr
		}

		// make service call
		protoResponse, err := info.method(info.svc.svc, ctx, dec, info.svc.interceptors)

		//apply decoder if any
		if info.decoder != nil {
			info.decoder(ctx, resp, encErr, err, protoResponse)
			if encErr != nil {
				return ctx, encErr
			}
			return ctx, err
		} else if dec, ok := h.defDecoders[cleanSvcName(info.svc.desc.ServiceName)]; ok {
			dec(ctx, resp, encErr, err, protoResponse)
			if encErr != nil {
				return ctx, encErr
			}
			return ctx, err
		}

		hdr := headers.ResponseHeadersFromContext(ctx)
		responseHeaders := processWhitelist(hdr, append(info.svc.responseHeaders, DefaultHTTPResponseHeaders...))
		if err != nil {
			if encErr != nil {
				writeRespWithHeaders(resp, http.StatusBadRequest, []byte("Bad Request!"), responseHeaders)
				return ctx, fmt.Errorf("Bad Request")
			}
			code, msg := GrpcErrorToHTTP(err, http.StatusInternalServerError, "Internal Server Error!")
			writeRespWithHeaders(resp, code, []byte(msg), responseHeaders)
			return ctx, fmt.Errorf(msg)
		}
		return ctx, h.serializeOut(ctx, resp, protoResponse.(proto.Message), responseHeaders)
	}
	writeResp(resp, http.StatusNotFound, []byte("Not Found: "+url))
	return req.Context(), fmt.Errorf("Not Found: " + url)
}

func (h *httpHandler) serialize(ctx context.Context, msg proto.Message) ([]byte, string, error) {
	serType, _ := modifiers.GetSerialization(ctx)
	if serType == "" {
		serType = ContentTypeFromHeaders(ctx)
		if serType == "" {
			serType = modifiers.JSON
		}
	}
	switch serType {
	case modifiers.JSONPB:
		sData, err := h.mar.MarshalToString(msg)
		return []byte(sData), "application/json", err
	case modifiers.ProtoBuf:
		data, err := proto.Marshal(msg)
		return data, "application/octet-stream", err
	case modifiers.JSON:
		fallthrough
	default:
		// modifiers.JSON goes in here
		data, err := json.Marshal(msg)
		return data, "application/json", err
	}
}

func (h *httpHandler) serializeOut(ctx context.Context, resp http.ResponseWriter, msg proto.Message, responseHeaders http.Header) error {
	data, contentType, err := h.serialize(ctx, msg)
	if err != nil {
		writeRespWithHeaders(resp, http.StatusInternalServerError, []byte("Internal Server Error!"), responseHeaders)
		return fmt.Errorf("Internal Server Error")
	}
	responseHeaders.Add("Content-Type", contentType)
	writeRespWithHeaders(resp, http.StatusOK, data, responseHeaders)
	return nil
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

	svcInfo.interceptors = getInterceptors(ss, h.config.CommonConfig)
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

		if h.config.EnableProtoURL {
			// add proto urls if enabled
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
			if strings.TrimSpace(path) != "" {
				info.encoderPath = path
				h.paths[path] = info
			} else {
				info.encoderPath = url
			}
		} else {
			fmt.Println("url not found", url, h.paths)
		}
	}
}

func (h *httpHandler) AddDefaultEncoder(serviceName string, encoder Encoder) {
	if h.defEncoders == nil {
		h.defEncoders = make(map[string]Encoder)
	}
	if encoder != nil {
		h.defEncoders[cleanSvcName(serviceName)] = encoder
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

func (h *httpHandler) AddDefaultDecoder(serviceName string, decoder Decoder) {
	if h.defDecoders == nil {
		h.defDecoders = make(map[string]Decoder)
	}
	if decoder != nil {
		h.defDecoders[cleanSvcName(serviceName)] = decoder
	}
}

func (h *httpHandler) Run(httpListener net.Listener) error {
	h.mu.Lock()
	r := mux.NewRouter()
	fmt.Println("Mapped URLs: ")
	for url, info := range h.paths {
		if strings.TrimSpace(info.encoderPath) != "" && info.encoderPath != url {
			continue
		}
		routeURL := url
		r.Methods(info.httpMethod...).Path(url).Handler(h.getHTTPHandler(url))
		if !strings.HasSuffix(url, "/") {
			routeURL = url + "/"
			r.Methods(info.httpMethod...).Path(url + "/").Handler(h.getHTTPHandler(url))
		}
		fmt.Println("\t", info.httpMethod, routeURL)
	}
	h.svr = &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      r,
	}
	s := h.svr
	h.mu.Unlock()
	return s.Serve(httpListener)
}

func (h *httpHandler) Stop(timeout time.Duration) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	ctx, can := context.WithTimeout(context.Background(), timeout)
	defer can()
	h.svr.Shutdown(ctx)
	return nil
}
