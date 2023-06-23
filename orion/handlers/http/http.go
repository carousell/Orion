package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/textproto"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	opentracing "github.com/opentracing/opentracing-go"
	"google.golang.org/grpc/metadata"

	"github.com/carousell/Orion/orion/handlers"
	"github.com/carousell/Orion/orion/modifiers"
	"github.com/carousell/Orion/utils"
	"github.com/carousell/Orion/utils/errors"
	"github.com/carousell/Orion/utils/errors/notifier"
	"github.com/carousell/Orion/utils/headers"
	"github.com/carousell/Orion/utils/log"
	"github.com/carousell/Orion/utils/log/loggers"
	"github.com/carousell/Orion/utils/options"
)

// grpcMetadataCarrier satisfies both opentracing.TextMapWriter and opentracing.TextMapReader.
type grpcMetadataCarrier metadata.MD

// ForeachKey conforms to the opentracing.TextMapReader interface.
func (mc grpcMetadataCarrier) ForeachKey(handler func(string, string) error) (err error) {
	for key, values := range mc {
		for _, value := range values {
			if err = handler(key, value); err != nil {
				return err
			}
		}
	}
	return nil
}

// Set conforms to the opentracing.TextMapWriter interface.
// Using this carrier ensures metadata keys are lowercased to conform to the
// HTTP/2 spec.
func (mc grpcMetadataCarrier) Set(key, value string) {
	k := strings.ToLower(key)
	mc[k] = append(mc[k], value)
}

func (h *httpHandler) getHTTPHandler(serviceName, methodName, routeURL string) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		fmt.Println("MANNY - handling request")
		h.httpHandler(resp, req, serviceName, methodName, routeURL)
		h := resp.Header().Clone()
		fmt.Println("MANNY - response headers")
		for k, v := range h {
			fmt.Println("MANNY - RESPONSE header -", k, ":", v)
		}
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

		// Add HTTP response code as tag
		var tags notifier.Tags
		if err != nil {
			httpCode, _ := GrpcErrorToHTTP(err, http.StatusInternalServerError, "Internal Server Error!")
			tags = notifier.Tags{
				"http_code": strconv.Itoa(httpCode),
			}
		}

		notifier.Notify(err, req.URL.String(), ctx, tags)
		utils.FinishNRTransaction(req.Context(), err)
	}
}

func (h *httpHandler) getNRTxName(req *http.Request, service, method, routeURL string) string {
	var nrTxName string
	switch h.config.NRHttpTxNameType {
	case NRTxNameTypeMethod:
		nrTxName = method
	case NRTxNameTypeRoute:
		nrTxName = fmt.Sprintf("%v %v", req.Method, routeURL)
	default:
		nrTxName = fmt.Sprintf("%v/%v", service, method)
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
			if values, found := req.Header[textproto.CanonicalMIMEHeaderKey(hdr)]; found {
				value := ""
				if len(values) > 0 {
					value = values[0]
				}
				ctx = headers.AddToRequestHeaders(ctx, hdr, value)
			}
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
		opentracing.GlobalTracer().Inject(wireContext, opentracing.HTTPHeaders, grpcMetadataCarrier(md))
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
			fmt.Println("SALLY - applying endpoint decoder")
			info.decoder(ctx, resp, encErr, err, protoResponse)
			if encErr != nil {
				return ctx, encErr
			}

			h := headers.ResponseHeadersFromContext(ctx)
			for k, v := range h {
				fmt.Println("SALLY - response header", k, ":", v)
			}
			return ctx, err
		} else if dec, ok := h.defDecoders[cleanSvcName(info.svc.desc.ServiceName)]; ok {
			fmt.Println("SALLY - applying service-level decoder")
			dec(ctx, resp, encErr, err, protoResponse)
			if encErr != nil {
				return ctx, encErr
			}

			h := headers.ResponseHeadersFromContext(ctx)
			for k, v := range h {
				fmt.Println("SALLY - response header", k, ":", v)
			}
			return ctx, err
		}
		fmt.Println("SALLY - applying default decoder")

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

	fmt.Println("BOBBY - get from orion modifiers -", serType)
	if serType == "" {
		// try and match an accept header
		serType = AcceptTypeFromHeaders(ctx)
		fmt.Println("BOBBY - get from accept type -", serType)
		if serType == "" {
			// try and match the original content type
			serType = ContentTypeFromHeaders(ctx)
			fmt.Println("BOBBY - get from content type -", serType)
			if serType == "" {

				fmt.Println("BOBBY - Default to json -", serType)
				serType = modifiers.JSON
			}
		}
		// if server preference is JSONPB, JSONPB should be used instead of JSON for marshalling
		if serType == modifiers.JSON && h.config.DefaultJSONPB {
			serType = modifiers.JSONPB
			fmt.Println("BOBBY - Default to jsonpb -", serType)
		}
	}
	fmt.Println("BOBBY - Final -", serType)
	switch serType {
	case modifiers.JSONPB:
		fmt.Println("LARRY - serializing as PROTOBUF (JSONPB")
		sData, err := h.mar.MarshalToString(msg)
		return []byte(sData), ContentTypeJSON, err
	case modifiers.ProtoBuf:
		fmt.Println("LARRY - serializing as PROTOBUF (WIRE)")
		data, err := proto.Marshal(msg)
		return data, ContentTypeProto, err
	case modifiers.JSON:
		fallthrough
	default:
		fmt.Println("LARRY - serializing as json")
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
	fmt.Println("DANNY - Writing contentType header", contentType)

	fmt.Printf("DANNY - resp writer type - %T\n", resp)
	writeRespWithHeaders(resp, http.StatusOK, data, responseHeaders)
	return nil
}
