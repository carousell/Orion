package http

import (
	"context"
	"net/http"
	"strings"

	"github.com/carousell/Orion/utils/errors"
	"github.com/carousell/Orion/utils/headers"
	"github.com/carousell/Orion/utils/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//ContentTypeFromHeaders searches for a matching content type
func ContentTypeFromHeaders(ctx context.Context) string {
	hdrs := headers.RequestHeadersFromContext(ctx)
	if values, found := hdrs["Content-Type"]; found {
		for _, v := range values {
			if t, ok := ContentTypeMap[v]; ok {
				return t
			}
		}
	}
	return ""
}

//AcceptTypeFromHeaders searches for a mathing accept type
func AcceptTypeFromHeaders(ctx context.Context) string {
	hdrs := headers.RequestHeadersFromContext(ctx)
	if values, found := hdrs["Accept"]; found {
		for _, v := range values {
			if t, ok := ContentTypeMap[v]; ok {
				return t
			}
		}
	}
	return ""
}

// GrpcErrorToHTTP converts gRPC error code into HTTP response status code.
// See: https://github.com/googleapis/googleapis/blob/master/google/rpc/code.proto
func GrpcErrorToHTTP(err error, defaultStatus int, defaultMessage string) (int, string) {
	httpStatus := defaultStatus
	msg := defaultMessage
	var code codes.Code

	if s, ok := status.FromError(err); ok {
		msg = s.Message()
		code = s.Code()
	} else if g, ok := err.(errors.GRPCExt); ok {
		code = g.Code()
		msg = code.String()
	}
	switch code {
	case codes.NotFound:
		httpStatus = http.StatusNotFound
	case codes.InvalidArgument:
		httpStatus = http.StatusBadRequest
	case codes.Unauthenticated:
		httpStatus = http.StatusUnauthorized
	case codes.PermissionDenied:
		httpStatus = http.StatusForbidden
	case codes.OK:
		httpStatus = http.StatusOK
	case codes.Canceled:
		httpStatus = http.StatusRequestTimeout
	case codes.Unknown:
		httpStatus = http.StatusInternalServerError
	case codes.DeadlineExceeded:
		httpStatus = http.StatusGatewayTimeout
	case codes.AlreadyExists:
		httpStatus = http.StatusConflict
	case codes.ResourceExhausted:
		httpStatus = http.StatusTooManyRequests
	case codes.FailedPrecondition:
		httpStatus = http.StatusBadRequest
	case codes.Aborted:
		httpStatus = http.StatusConflict
	case codes.OutOfRange:
		httpStatus = http.StatusBadRequest
	case codes.Unimplemented:
		httpStatus = http.StatusNotImplemented
	case codes.Internal:
		httpStatus = http.StatusInternalServerError
	case codes.Unavailable:
		httpStatus = http.StatusServiceUnavailable
	case codes.DataLoss:
		httpStatus = http.StatusInternalServerError
	}

	return httpStatus, msg
}

func processWhitelist(ctx context.Context, data map[string][]string, allowedKeys []string) map[string][]string {
	whitelistedMap := make(map[string][]string)
	whitelistedKeys := make(map[string]bool)

	for _, k := range allowedKeys {
		whitelistedKeys[strings.ToLower(k)] = true
	}

	for k, v := range data {
		if _, found := whitelistedKeys[strings.ToLower(k)]; found {
			whitelistedMap[k] = v
		} else {
			log.Warn(ctx, "error", "rejected headers not in whitelist", k, v)
		}
	}

	return whitelistedMap
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
