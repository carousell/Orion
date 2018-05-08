package notifier

import (
	"context"
	"log"
	"os"
	"reflect"
	"strings"

	bugsnag "github.com/bugsnag/bugsnag-go"
	"github.com/carousell/Orion/utils/errors"
	"github.com/carousell/Orion/utils/options"
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
	tracerID = "tracerId"
)

// InitAirbrake inits airbrake configuration
func InitAirbrake(projectID int64, projectKey string) {
	airbrake = gobrake.NewNotifier(projectID, projectKey)
}

//InitBugsnag inits bugsnag configuration
func InitBugsnag(config bugsnag.Configuration) {
	bugsnag.Configure(config)
	bugsnagInited = true
}

//InitRollbar inits rollbar configuration
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

	if n, ok := err.(errors.NotifyExt); ok {
		if !n.ShouldNotify() {
			return err
		}
		n.Notified(true)
	}
	return doNotify(err, skip, level, rawData...)

}

func doNotify(err error, skip int, level string, rawData ...interface{}) error {
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
	var traceID string
	for _, d := range list {
		if c, ok := d.(context.Context); ok {
			if span := stdopentracing.SpanFromContext(c); span != nil {
				traceID = span.BaggageItem("trace")
			}
			if strings.TrimSpace(traceID) == "" {
				traceID = GetTraceId(c)
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
		if traceID != "" {
			n.Context["traceId"] = traceID
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
		if traceID != "" {
			fields = append(fields, &rollbar.Field{Name: "traceId", Data: traceID})
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
			e := errors.NewWithSkip("Panic: ", 1)
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

func SetTraceId(ctx context.Context) context.Context {
	if GetTraceId(ctx) != "" {
		return ctx
	}
	var traceID string
	if span := stdopentracing.SpanFromContext(ctx); span != nil {
		traceID = span.BaggageItem("trace")
	}
	// if no trace id then create one
	if strings.TrimSpace(traceID) == "" {
		traceID = uuid.NewUUID().String()
	}
	return options.AddToOptions(ctx, tracerID, traceID)
}

func GetTraceId(ctx context.Context) string {
	if o := options.FromContext(ctx); o != nil {
		if data, found := o.Get(tracerID); found {
			return data.(string)
		}
	}
	return ""
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
