package utils

import (
	"context"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/carousell/go-utils/utils/errors/notifier"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	newrelic "github.com/newrelic/go-agent"
	stdopentracing "github.com/opentracing/opentracing-go"
	uuid "github.com/satori/go.uuid"
)

const (
	newRelicTransactionId string = "NewRelicTransaction"
	loggerId              string = "logger"
)

func GetNewRelicTransactionFromContext(ctx context.Context) newrelic.Transaction {
	t := ctx.Value(newRelicTransactionId)
	if t != nil {
		txn, ok := t.(newrelic.Transaction)
		if ok {
			return txn
		}
	}
	return nil
}

func StoreNewRelicTransactionToContext(ctx context.Context, t newrelic.Transaction) context.Context {
	return context.WithValue(ctx, newRelicTransactionId, t)
}

func StoreLoggerToContext(ctx context.Context, logger log.Logger) context.Context {
	return context.WithValue(ctx, loggerId, logger)
}

func GetLoggerFromContext(ctx context.Context) log.Logger {
	l := ctx.Value(loggerId)
	if l != nil {
		logger, ok := l.(log.Logger)
		if ok {
			return logger
		}
	}
	return nil
}

func UpdateHTTPTracingSpanWithTag(f httptransport.DecodeRequestFunc, tracerKey string) httptransport.DecodeRequestFunc {
	return func(ctx context.Context, r *http.Request) (request interface{}, err error) {
		if span := stdopentracing.SpanFromContext(ctx); span != nil {
			if value := r.Header.Get(tracerKey); value != "" {
				span.SetTag("trace", value)
				span.SetBaggageItem("trace", value)
				exTrace := notifier.GetTraceId(ctx)
				if exTrace != "" {
					span.SetTag("exTrace", exTrace)
				}
			}
		}
		return f(ctx, r)
	}
}

type ErrorChecker func(error) bool

func NewRelicMiddleware(name string, app newrelic.Application) endpoint.Middleware {
	return NewRelicMiddlewareWithErrorCheck(name, app, nil)
}

func NewRelicMiddlewareWithErrorCheck(name string, app newrelic.Application, check ErrorChecker) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			req, _ := http.NewRequest("", "/"+name, nil)
			if app != nil {
				t := app.StartTransaction(name, nil, req)
				ctx = StoreNewRelicTransactionToContext(ctx, t)
				response, err = next(ctx, request)
				if check != nil && err != nil {
					// check is defined
					if check(err) {
						t.NoticeError(err)
					}
				} else {
					t.NoticeError(err)
				}
				t.End()
			} else {
				response, err = next(ctx, request)
			}
			return response, err
		}
	}
}

// EndpointLoggingMiddleware returns an endpoint middleware that logs the
// duration of each invocation, and the resulting error, if any.
func EndpointLoggingMiddleware(logger log.Logger) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			ctx = StoreLoggerToContext(ctx, logger)
			defer func(begin time.Time) {
				logger.Log("error", err, "took", time.Since(begin))
			}(time.Now())
			if next != nil {
				return next(ctx, request)
			} else {
				logger.Log("Error", "got null endpoint", "req", request)
				return response, errors.New("null endpoint recieved by EndpointLoggingMiddleware")
			}
		}
	}
}

//ignore case true would not covert the strings to lower case, hence ignoring the case.
func Contains(s []string, e string, ignoreCase bool) bool {

	if ignoreCase {
		e = strings.ToLower(e)
		for i, v := range s {
			s[i] = strings.ToLower(v)
		}
	}
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func CleanUUID(uuidStr string) (string, error) {
	u, err := uuid.FromString(uuidStr)
	if err == nil {
		return u.String(), nil
	}
	bytes, err := hex.DecodeString(uuidStr)
	if err != nil {
		return uuidStr, err
	}
	u, err = uuid.FromBytes(bytes)
	if err != nil {
		return uuidStr, err
	}
	return u.String(), nil
}
