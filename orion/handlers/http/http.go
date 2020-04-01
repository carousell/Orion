package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/carousell/Orion/orion/handlers"
	"github.com/carousell/Orion/orion/modifiers"
	"github.com/carousell/Orion/utils"
	"github.com/carousell/Orion/utils/errors"
	"github.com/carousell/Orion/utils/errors/notifier"
	"github.com/carousell/Orion/utils/headers"
	"github.com/carousell/Orion/utils/log"
	"github.com/carousell/Orion/utils/log/loggers"
	"github.com/carousell/Orion/utils/options"
	"github.com/golang/protobuf/proto"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	opentracing "github.com/opentracing/opentracing-go"
)

func (h *httpHandler) getHTTPHandler(serviceName, methodName, routeURL string) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		h.httpHandler(resp, req, serviceName, methodName, routeURL)
	}
}

func (h *httpHandler) httpHandler(resp http.ResponseWriter, req *http.Request, service, method, routeURL string) {
	nrTxName := h.getNRTxName(req, service, method, routeURL)
	ctx := utils.StartNRTransaction(nrTxName, req.Context(), req, resp)
	ctx = loggers.AddToLogContext(ctx, "transport", "http")
	var err error
	defer func(ctx context.Context, t time.Time) {
		log.Info(ctx, "path", req.URL.Path, "method", req.Method, "error", err, "took", time.Since(t))
	}(ctx, time.Now())
	req = req.WithContext(ctx)
	ctx, err = h.serveHTTP(resp, req, service, method)
	if modifiers.HasDontLogError(ctx) {
		utils.FinishNRTransaction(req.Context(), nil)
	} else {
		notifier.Notify(err, req.URL.String(), ctx)
		utils.FinishNRTransaction(req.Context(), err)
	}
}

func (h *httpHandler)  getNRTxName(req *http.Request, service, method, routeURL string) string {
	nrTxName := method
	switch h.config.NRHttpTxName {
	case NRTxNameFullMethod:
		nrTxName = fmt.Sprintf("%v/%v", service, method)
	case NRTxNameURL:
		nrTxName = fmt.Sprintf("%v %v", req.Method, req.URL)
	case NRTxNameRoute:
		nrTxName = fmt.Sprintf("%v %v", req.Method, routeURL)
	}
	return nrTxName
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

func (h *httpHandler) serveHTTP(resp http.ResponseWriter, req *http.Request, serviceName, methodName string) (context.Context, error) {
	info, ok := h.mapping.Get(serviceName, methodName)
	if ok {
		ctx := prepareContext(req, info)
		ctx = processOptions(ctx, req, info)
		req = req.WithContext(ctx)
		// httpHandler allows handling entire http request
		if info.httpHandler != nil {
			if info.httpHandler(resp, req) {
				// short circuit if handler has handled request
				return ctx, nil
			}
		} else if handlerFetcher, ok := info.svc.svc.(handlers.CustomHTTPHandler); ok {
			h := handlerFetcher.GetHTTPHandler(strings.ToLower(info.methodName))
			if h != nil {
				if h(resp, req) {
					return ctx, nil
				}
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
				encErr = DefaultEncoder(req, r, h.config.DefaultJSONPB)
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
		responseHeaders := processWhitelist(ctx, hdr, append(info.svc.responseHeaders, DefaultHTTPResponseHeaders...))
		if err != nil {
			if encErr != nil {
				writeRespWithHeaders(resp, http.StatusBadRequest, []byte("Bad Request!"), responseHeaders)
				return ctx, errors.Wrap(encErr, "Bad Request")
			}
			code, msg := GrpcErrorToHTTP(err, http.StatusInternalServerError, "Internal Server Error!")
			writeRespWithHeaders(resp, code, []byte(msg), responseHeaders)
			return ctx, errors.Wrap(err, msg)
		}
		return ctx, h.serializeOut(ctx, resp, protoResponse.(proto.Message), responseHeaders)
	}
	writeResp(resp, http.StatusNotFound, []byte("Not Found: "+req.URL.String()))
	return req.Context(), errors.New("Not Found: " + req.URL.String())
}

func (h *httpHandler) serialize(ctx context.Context, msg proto.Message) ([]byte, string, error) {
	// first check if any serialization is
	serType, _ := modifiers.GetSerialization(ctx)
	if serType == "" {
		// try and match an accept header
		serType = AcceptTypeFromHeaders(ctx)
		if serType == "" {
			// try and match the original content type
			serType = ContentTypeFromHeaders(ctx)
			if serType == "" {
				serType = modifiers.JSON
			}
		}
		// if server preference is JSONPB, JSONPB should be used instead of JSON for marshalling
		if serType == modifiers.JSON && h.config.DefaultJSONPB {
			serType = modifiers.JSONPB
		}
	}
	switch serType {
	case modifiers.JSONPB:
		sData, err := h.mar.MarshalToString(msg)
		return []byte(sData), ContentTypeJSON, err
	case modifiers.ProtoBuf:
		data, err := proto.Marshal(msg)
		return data, ContentTypeProto, err
	case modifiers.JSON:
		fallthrough
	default:
		// modifiers.JSON goes in here
		data, err := json.Marshal(msg)
		return data, ContentTypeJSON, err
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
