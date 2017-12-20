package notifier

import (
	"context"
	"log"
	"os"
	"reflect"

	bugsnag "github.com/bugsnag/bugsnag-go"
	"github.com/carousell/go-utils/utils/errors"
	"github.com/go-kit/kit/endpoint"
	stdopentracing "github.com/opentracing/opentracing-go"
	"github.com/pborman/uuid"
	"github.com/stvp/rollbar"
	gobrake "gopkg.in/airbrake/gobrake.v2"
)

var (
	airbrake      *gobrake.Notifier
	bugsnagInited bool
	rollbarInited bool
	serverRoot    string
	hostname      string
)

const (
	tracerId string = "tracerId"
)

type MetaData map[string]interface{}

func InitAirbrake(projectId int64, projectKey string) {
	airbrake = gobrake.NewNotifier(projectId, projectKey)
}

func InitBugsnag(config bugsnag.Configuration) {
	bugsnag.Configure(config)
	bugsnagInited = true
}

func InitRollbar(token, env string) {
	rollbar.Token = token
	rollbar.Environment = env
	rollbarInited = true
}

func convToGoBrake(in []errors.StackFrame) []gobrake.StackFrame {
	out := make([]gobrake.StackFrame, 0)
	for _, s := range in {
		out = append(out, gobrake.StackFrame{
			File: s.File,
			Func: s.Func,
			Line: s.Line,
		})
	}
	return out
}

func convToRollbar(in []errors.StackFrame) rollbar.Stack {
	out := rollbar.Stack{}
	for _, s := range in {
		out = append(out, rollbar.Frame{
			Filename: s.File,
			Method:   s.Func,
			Line:     s.Line,
		})
	}
	return out
}

func parseRawData(rawData ...interface{}) map[string]interface{} {
	m := make(map[string]interface{})
	for pos := range rawData {
		data := rawData[pos]
		if _, ok := data.(context.Context); ok {
			continue
		}
		m[reflect.TypeOf(data).String()] = data
	}
	return m
}

func Notify(err error, rawData ...interface{}) error {
	return NotifyWithLevelAndSkip(err, 2, rollbar.ERR, rawData...)
}

func NotifyWithLevel(err error, level string, rawData ...interface{}) error {
	return NotifyWithLevelAndSkip(err, 2, level, rawData...)
}

func NotifyWithLevelAndSkip(err error, skip int, level string, rawData ...interface{}) error {
	if err == nil {
		return nil
	}

	list := make([]interface{}, 0)
	for pos := range rawData {
		data := rawData[pos]
		// if we find the error, return error and do not log it
		if e, ok := data.(error); ok {
			if e == err {
				return err
			} else if er, ok := e.(errors.ErrorExt); ok {
				if err == er.Cause() {
					return err
				}
			}
		} else {
			list = append(list, rawData[pos])
		}
	}
	var traceId string
	for _, d := range list {
		if c, ok := d.(context.Context); ok {
			if span := stdopentracing.SpanFromContext(c); span != nil {
				traceId = span.BaggageItem("trace")
			} else {
				traceId = GetTraceId(c)
			}
			break
		}
	}
	if airbrake != nil {
		var n *gobrake.Notice
		if e, ok := err.(errors.ErrorExt); ok {
			// airbrake needs different format for stackframe
			n = gobrake.NewNotice(e, nil, 500)
			n.Errors[0].Backtrace = convToGoBrake(e.StackFrame())
		} else {
			n = gobrake.NewNotice(e, nil, skip)
		}
		if len(list) > 0 {
			m := parseRawData(list...)
			for k, v := range m {
				n.Context[k] = v
			}
		}
		if traceId != "" {
			n.Context["traceId"] = traceId
		}
		airbrake.SendNoticeAsync(n)
	}
	if bugsnagInited {
		bugsnag.Notify(err, list...)
	}
	if rollbarInited {
		fields := []*rollbar.Field{}
		if len(list) > 0 {
			m := parseRawData(list...)
			for k, v := range m {
				fields = append(fields, &rollbar.Field{Name: k, Data: v})
			}
		}
		if traceId != "" {
			fields = append(fields, &rollbar.Field{Name: "traceId", Data: traceId})
		}
		fields = append(fields, &rollbar.Field{Name: "server", Data: map[string]interface{}{"hostname": getHostname(), "root": getServerRoot()}})
		if e, ok := err.(errors.ErrorExt); ok {
			// rollbar needs different format for stackframe
			rollbar.ErrorWithStack(level, e, convToRollbar(e.StackFrame()), fields...)
			log.Println(level, e, e.StackFrame())
		} else {
			e := errors.WrapWithSkip(err, "", skip)
			log.Println(level, e, e.StackFrame())
			rollbar.ErrorWithStack(level, e, convToRollbar(e.StackFrame()), fields...)
		}
	}
	return err
}

func NotifyWithExclude(err error, rawData ...interface{}) error {
	if err == nil {
		return nil
	}

	list := make([]interface{}, 0)
	for pos := range rawData {
		data := rawData[pos]
		// if we find the error, return error and do not log it
		if e, ok := data.(error); ok {
			if er, ok := e.(errors.ErrorExt); ok {
				if err == er.Cause() {
					return err
				} else if er == err {
					return err
				}
			}
		} else {
			list = append(list, rawData[pos])
		}
	}
	go Notify(err, list...)
	return err
}

func NotifyOnPanic(rawData ...interface{}) {
	if bugsnagInited {
		bugsnag.AutoNotify(rawData...)
	}
	if airbrake != nil {
		airbrake.NotifyOnPanic()
	}
	if rollbarInited {
		if r := recover(); r != nil {
			var e errors.ErrorExt
			//if _, ok := r.(error); ok {
			//	e = errors.WrapWithSkip(e, "Panic ", 1)
			//} else {
			e = errors.NewWithSkip("Panic: ", 1)
			//}
			rollbar.ErrorWithStack(rollbar.CRIT, e, convToRollbar(e.StackFrame()), &rollbar.Field{Name: "panic", Data: r})
			panic(r)
		}
	}
}

func Close() {
	if airbrake != nil {
		airbrake.Close()
	}
}

func SetEnvironemnt(env string) {
	if airbrake != nil {
		airbrake.AddFilter(func(notice *gobrake.Notice) *gobrake.Notice {
			notice.Context["environment"] = env
			return notice
		})
	}
	rollbar.Environment = env
}

func NotifyingMiddleware(name string) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			m := MetaData{
				"Name": name,
			}
			defer NotifyOnPanic()
			response, err = next(ctx, request)
			if err != nil {
				Notify(err, m, ctx, request)
			}
			return response, err
		}
	}
}

func SetTraceId(ctx context.Context) context.Context {
	if GetTraceId(ctx) != "" {
		return ctx
	}
	var traceId string
	if span := stdopentracing.SpanFromContext(ctx); span != nil {
		traceId = span.BaggageItem("trace")
	} else {
		traceId = uuid.NewUUID().String()
	}
	return context.WithValue(ctx, tracerId, traceId)
}

func GetTraceId(ctx context.Context) string {
	data := ctx.Value(tracerId)
	if data == nil {
		return ""
	}
	v, _ := data.(string)
	return v
}

func TracingMiddleware() endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			ctx = SetTraceId(ctx)
			response, err = next(ctx, request)
			return response, err
		}
	}
}

func SetServerRoot(path string) {
	serverRoot = path
}

func SetHostname(name string) {
	hostname = name
}

func getHostname() string {
	if hostname != "" {
		return hostname
	}
	name, err := os.Hostname()
	if err == nil {
		hostname = name
	}
	return hostname
}

func getServerRoot() string {
	if serverRoot != "" {
		return serverRoot
	}
	cwd, err := os.Getwd()
	if err == nil {
		serverRoot = cwd
	}
	return serverRoot
}
