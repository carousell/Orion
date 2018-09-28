package http

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/carousell/Orion/orion/handlers"
	"github.com/carousell/Orion/utils/log"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
)

//NewHTTPHandler creates a new HTTP handler
func NewHTTPHandler(config Config) handlers.Handler {
	return &httpHandler{
		config:      config,
		mapping:     newMethodInfoMapping(),
		middlewares: handlers.NewMiddlewareMapping(),
	}
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
	for _, s := range sd.Streams {
		info := &methodInfo{
			stream:        s.Handler,
			svc:           svcInfo,
			httpMethod:    []string{"GET"},
			serviceName:   sd.ServiceName,
			methodName:    s.StreamName,
			urls:          make([]string, 0),
			clientStreams: s.ClientStreams,
			serverStreams: s.ServerStreams,
		}
		info.urls = append(info.urls, generateURL(sd.ServiceName, s.StreamName))
		if h.config.EnableProtoURL {
			// add proto urls if enabled
			info.urls = append(info.urls, generateProtoURL(sd.ServiceName, s.StreamName))
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
			log.Warn(context.Background(), "error", "Service and Method NOT found!", "service", serviceName,
				"method", method, "mapping", h.mapping)
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
			log.Warn(context.Background(), "error", "Service and Method NOT found!", "service", serviceName,
				"method", method, "mapping", h.mapping)
		}
	}
}

func (h *httpHandler) AddDecoder(serviceName, method string, decoder handlers.Decoder) {
	if h.mapping != nil {
		if info, ok := h.mapping.Get(serviceName, method); ok {
			info.decoder = decoder
		} else {
			log.Warn(context.Background(), "error", "Service and Method NOT found!", "service", serviceName,
				"method", method, "mapping", h.mapping)
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
			if strings.TrimSpace(info.encoderPath) != "" {
				if info.encoderPath != url {
					// only add the encoder url if encoder is defined, skip others
					continue
				}
			}
			routeURL := url
			var handler http.HandlerFunc
			methodClassifier := make([]string, 0)
			if info.clientStreams || info.serverStreams {
				handler = h.getWSHandler(info.serviceName, info.methodName)
				if info.clientStreams {
					methodClassifier = append(methodClassifier, "CLIENT_STREAMING")
				}
				if info.serverStreams {
					methodClassifier = append(methodClassifier, "SERVER_STREAMING")
				}
			} else {
				handler = h.getHTTPHandler(info.serviceName, info.methodName)
				methodClassifier = append(methodClassifier, "NON_STREAMING")
			}
			r.Methods(info.httpMethod...).Path(url).Handler(handler)
			if !strings.HasSuffix(url, "/") {
				routeURL = url + "/"
				r.Methods(info.httpMethod...).Path(url + "/").Handler(handler)
			}
			fmt.Println("\t", info.httpMethod, routeURL, "mapped to", info.serviceName, info.methodName, methodClassifier)
		}
	}
	r.NotFoundHandler = &notFoundHandler{}
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
