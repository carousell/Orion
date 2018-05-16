package http

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/carousell/Orion/utils/headers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//ContentTypeFromHeaders searches for a matching content type
func ContentTypeFromHeaders(ctx context.Context) string {
	hdrs := headers.RequestHeadersFromContext(ctx)
	if values, found := hdrs["Accept"]; found {
		for _, v := range values {
			if t, ok := ContentTypeMap[v]; ok {
				return t
			}
		}
	}
	if values, found := hdrs["Content-Type"]; found {
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

func processWhitelist(data map[string][]string, allowedKeys []string) map[string][]string {
	whitelistedMap := make(map[string][]string)
	whitelistedKeys := make(map[string]bool)

	for _, k := range allowedKeys {
		whitelistedKeys[strings.ToLower(k)] = true
	}

	for k, v := range data {
		if _, found := whitelistedKeys[strings.ToLower(k)]; found {
			whitelistedMap[k] = v
		} else {
			log.Println("warning", "rejected headers not in whitelist", k, v)
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
