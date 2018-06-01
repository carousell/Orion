package http

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
	"time"

	"github.com/carousell/Orion/orion/handlers"
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
)

//NewHTTPHandler creates a new HTTP handler
func NewHTTPHandler(config HandlerConfig) handlers.Handler {
	return &httpHandler{
		config:      config,
		mapping:     newMethodInfoMapping(),
		middlewares: handlers.NewMiddlewareMapping(),
	}
}

func (h *httpHandler) getHTTPHandler(serviceName, methodName string) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		h.httpHandler(resp, req, serviceName, methodName)
	}
}

func (h *httpHandler) httpHandler(resp http.ResponseWriter, req *http.Request, service, method string) {
	var err error
	ctx := utils.StartNRTransaction(req.URL.String(), req.Context(), req, resp)
	defer func(resp http.ResponseWriter, ctx context.Context, t time.Time) {
		// panic handler
		if r := recover(); r != nil {
			writeResp(resp, http.StatusInternalServerError, []byte("Internal Server Error!"))
			log.Println("panic", r, "path", req.URL.String(), "method", req.Method, "took", time.Since(t))
			log.Print(string(debug.Stack()))
			var err error
			if e, ok := r.(error); ok {
				err = e
			} else {
				err = fmt.Errorf("panic: %s", r)
			}
			utils.FinishNRTransaction(ctx, err)
			notifier.NotifyWithLevel(err, "critical", req.URL.String(), ctx)
		} else {
			log.Println("path", req.URL.String(), "method", req.Method, "error", err, "took", time.Since(t))
		}
	}(resp, ctx, time.Now())
	req = req.WithContext(ctx)
	ctx, err = h.serveHTTP(resp, req, service, method)
	if modifiers.HasDontLogError(ctx) {
		utils.FinishNRTransaction(req.Context(), nil)
	} else {
		notifier.Notify(err, req.URL.String(), ctx)
		utils.FinishNRTransaction(req.Context(), err)
	}
}

func prepareContext(req *http.Request, info *methodInfo) context.Context {
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

func processOptions(ctx context.Context, req *http.Request, info *methodInfo) context.Context {
	if info.options != nil {
		for _, opt := range info.options {
			switch strings.ToUpper(opt) {
			case IgnoreNR:
				utils.IgnoreNRTransaction(ctx)
			}
		}
	}
	return ctx
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

func (h *httpHandler) serveHTTP(resp http.ResponseWriter, req *http.Request, serviceName, methodName string) (context.Context, error) {
	info, ok := h.mapping.Get(serviceName, methodName)
	if ok {
		ctx := prepareContext(req, info)
		ctx = processOptions(ctx, req, info)
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

		// fetch all method middlewares
		middlewares := make([]string, 0)
		if h.middlewares != nil {
			middlewares = append(middlewares, h.middlewares.GetMiddlewares(info.serviceName, info.methodName)...)
		}
		// fetch all interceptors
		interceptors := handlers.GetInterceptorsWithMethodMiddlewares(info.svc.svc, h.config.CommonConfig, middlewares)

		// make service call
		protoResponse, err := info.method(info.svc.svc, ctx, dec, interceptors)

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
	writeResp(resp, http.StatusNotFound, []byte("Not Found: "+req.URL.String()))
	return req.Context(), fmt.Errorf("Not Found: " + req.URL.String())
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

	svcInfo := &serviceInfo{
		desc: sd,
		svc:  ss,
	}

	if headers, ok := ss.(handlers.WhitelistedHeaders); ok {
		svcInfo.requestHeaders = headers.GetRequestHeaders()
		svcInfo.responseHeaders = headers.GetResponseHeaders()
	}

	// TODO recover in case of error
	for _, m := range sd.Methods {
		info := &methodInfo{
			method:      handlers.GRPCMethodHandler(m.Handler),
			svc:         svcInfo,
			httpMethod:  []string{"POST"},
			serviceName: sd.ServiceName,
			methodName:  m.MethodName,
			urls:        make([]string, 0),
		}
		info.urls = append(info.urls, generateURL(sd.ServiceName, m.MethodName))
		if h.config.EnableProtoURL {
			// add proto urls if enabled
			info.urls = append(info.urls, generateProtoURL(sd.ServiceName, m.MethodName))
		}
		h.mapping.Add(info.serviceName, info.methodName, info)
	}
	return nil
}

func (h *httpHandler) AddEncoder(serviceName, method string, httpMethod []string, path string, encoder handlers.Encoder) {
	if h.mapping != nil {
		if info, ok := h.mapping.Get(serviceName, method); ok {
			info.encoder = encoder
			info.httpMethod = httpMethod
			url := generateURL(serviceName, method)
			if strings.TrimSpace(path) != "" {
				info.encoderPath = path
				info.urls = append(info.urls, path)
			} else {
				info.encoderPath = url
			}
		} else {
			fmt.Println("Service and Method NOT found!", serviceName, method, h.mapping)
		}
	}
}

func (h *httpHandler) AddDefaultEncoder(serviceName string, encoder handlers.Encoder) {
	if h.defEncoders == nil {
		h.defEncoders = make(map[string]handlers.Encoder)
	}
	if encoder != nil {
		h.defEncoders[cleanSvcName(serviceName)] = encoder
	}
}

func (h *httpHandler) AddHTTPHandler(serviceName string, method string, path string, handler handlers.HTTPHandler) {
	if h.mapping != nil {
		if info, ok := h.mapping.Get(serviceName, method); ok {
			info.httpHandler = handler
		} else {
			fmt.Println("Service and Method NOT found!", serviceName, method, h.mapping)
		}
	}
}

func (h *httpHandler) AddDecoder(serviceName, method string, decoder handlers.Decoder) {
	if h.mapping != nil {
		if info, ok := h.mapping.Get(serviceName, method); ok {
			info.decoder = decoder
		} else {
			fmt.Println("Service and Method NOT found!", serviceName, method, h.mapping)
		}
	}
}

func (h *httpHandler) AddDefaultDecoder(serviceName string, decoder handlers.Decoder) {
	if h.defDecoders == nil {
		h.defDecoders = make(map[string]handlers.Decoder)
	}
	if decoder != nil {
		h.defDecoders[cleanSvcName(serviceName)] = decoder
	}
}

func (h *httpHandler) AddOption(serviceName, method, option string) {
	if info, ok := h.mapping.Get(serviceName, method); ok {
		if info.options == nil {
			info.options = make([]string, 0)
		}
		info.options = append(info.options, option)
	}

}

func (h *httpHandler) AddMiddleware(serviceName string, method string, middlewares ...string) {
	if h.middlewares == nil {
		h.middlewares = handlers.NewMiddlewareMapping()
	}
	h.middlewares.AddMiddleware(serviceName, method, middlewares...)
}

func (h *httpHandler) Run(httpListener net.Listener) error {
	r := mux.NewRouter()
	fmt.Println("Mapped URLs: ")
	allPaths := h.mapping.GetAllMethodInfoByOrder()
	for i := range allPaths {
		info := allPaths[i]
		for _, url := range info.urls {
			if strings.TrimSpace(info.encoderPath) != "" && info.encoderPath != url {
				// only add the encoder url if encoder is defined, skip others
				continue
			}
			routeURL := url
			handler := h.getHTTPHandler(info.serviceName, info.methodName)
			r.Methods(info.httpMethod...).Path(url).Handler(handler)
			if !strings.HasSuffix(url, "/") {
				routeURL = url + "/"
				r.Methods(info.httpMethod...).Path(url + "/").Handler(handler)
			}
			fmt.Println("\t", info.httpMethod, routeURL, "mapped to", info.serviceName, info.methodName)
		}
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
