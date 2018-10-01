package notifier

import (
	"context"
	"os"
	"reflect"
	"strconv"
	"strings"

	bugsnag "github.com/bugsnag/bugsnag-go"
	"github.com/carousell/Orion/utils/errors"

	"github.com/carousell/Orion/utils/log"
	"github.com/carousell/Orion/utils/log/loggers"
	"github.com/carousell/Orion/utils/options"
	raven "github.com/getsentry/raven-go"
	stdopentracing "github.com/opentracing/opentracing-go"
	"github.com/pborman/uuid"
	"github.com/stvp/rollbar"
	gobrake "gopkg.in/airbrake/gobrake.v2"
)

var (
	airbrake      *gobrake.Notifier
	bugsnagInited bool
	rollbarInited bool
	sentryInited  bool
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

func InitSentry(dsn string) {
	raven.SetDSN(dsn)
	sentryInited = true
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

func convToSentry(in []errors.StackFrame) *raven.Stacktrace {
	out := new(raven.Stacktrace)
	out.Frames = make([]*raven.StacktraceFrame, len(in))
	for i, s := range in {
		out.Frames[len(in)-i-1] = &raven.StacktraceFrame{
			Filename: s.File,
			Function: s.Func,
			Lineno:   s.Line,
		}
	}
	return out
}

func parseRawData(ctx context.Context, rawData ...interface{}) map[string]interface{} {
	m := make(map[string]interface{})
	for pos := range rawData {
		data := rawData[pos]
		if _, ok := data.(context.Context); ok {
			continue
		}
		m[reflect.TypeOf(data).String()+strconv.Itoa(pos)] = data
	}
	if logFields := loggers.FromContext(ctx); logFields != nil {
		for k, v := range logFields {
			m[k] = v
		}
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

	// add stack infomation
	errWithStack, ok := err.(errors.ErrorExt)
	if !ok {
		errWithStack = errors.WrapWithSkip(err, "", skip+1)
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

	// try to fetch a traceID and context from rawData
	var traceID string
	ctx := context.Background()
	for _, d := range list {
		if c, ok := d.(context.Context); ok {
			if span := stdopentracing.SpanFromContext(c); span != nil {
				traceID = span.BaggageItem("trace")
			}
			if strings.TrimSpace(traceID) == "" {
				traceID = GetTraceId(c)
			}
			ctx = c
			break
		}
	}

	if airbrake != nil {
		var n *gobrake.Notice
		n = gobrake.NewNotice(errWithStack, nil, 1)
		n.Errors[0].Backtrace = convToGoBrake(errWithStack.StackFrame())
		if len(list) > 0 {
			m := parseRawData(ctx, list...)
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
		bugsnag.Notify(errWithStack, list...)
	}
	parsedData := parseRawData(ctx, list...)
	if rollbarInited {
		fields := []*rollbar.Field{}
		if len(list) > 0 {
			for k, v := range parsedData {
				fields = append(fields, &rollbar.Field{Name: k, Data: v})
			}
		}
		if traceID != "" {
			fields = append(fields, &rollbar.Field{Name: "traceId", Data: traceID})
		}
		fields = append(fields, &rollbar.Field{Name: "server", Data: map[string]interface{}{"hostname": getHostname(), "root": getServerRoot()}})
		rollbar.ErrorWithStack(level, errWithStack, convToRollbar(errWithStack.StackFrame()), fields...)
	}

	if sentryInited {
		defLevel := raven.ERROR
		if level == "critical" {
			defLevel = raven.FATAL
		}
		ravenExp := raven.NewException(errWithStack, convToSentry(errWithStack.StackFrame()))
		packet := raven.NewPacketWithExtra(errWithStack.Error(), parsedData, ravenExp)
		packet.Level = defLevel
		raven.Capture(packet, nil)
	}

	log.GetLogger().Log(ctx, loggers.ErrorLevel, skip+1, "err", errWithStack, "stack", errWithStack.StackFrame())
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
		defer bugsnag.AutoNotify(rawData...)
	}
	if airbrake != nil {
		defer airbrake.NotifyOnPanic()
	}

	ctx := context.Background()
	for _, d := range rawData {
		if c, ok := d.(context.Context); ok {
			ctx = c
			break
		}
	}
	if r := recover(); r != nil {
		var e errors.ErrorExt
		switch val := r.(type) {
		case error:
			e = errors.WrapWithSkip(val, "PANIC", 1)
		case string:
			e = errors.NewWithSkip("PANIC: "+val, 1)
		default:
			e = errors.NewWithSkip("Panic", 1)
		}
		parsedData := parseRawData(ctx, rawData...)
		if rollbarInited {
			rollbar.ErrorWithStack(rollbar.CRIT, e, convToRollbar(e.StackFrame()), &rollbar.Field{Name: "panic", Data: r})
		}
		if sentryInited {
			ravenExp := raven.NewException(e, convToSentry(e.StackFrame()))
			packet := raven.NewPacketWithExtra(e.Error(), parsedData, ravenExp)
			packet.Level = raven.FATAL
			raven.Capture(packet, nil)
		}
		panic(e)
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
	raven.SetEnvironment(env)
}

//SetTraceId updates the traceID based on context values
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
	ctx = loggers.AddToLogContext(ctx, "trace", traceID)
	return options.AddToOptions(ctx, tracerID, traceID)
}

//GetTraceId fetches traceID from context
func GetTraceId(ctx context.Context) string {
	if o := options.FromContext(ctx); o != nil {
		if data, found := o.Get(tracerID); found {
			return data.(string)
		}
	}
	if logCtx := loggers.FromContext(ctx); logCtx != nil {
		if data, found := logCtx["trace"]; found {
			traceID := data.(string)
			options.AddToOptions(ctx, tracerID, traceID)
			return traceID
		}
	}
	return ""
}

//UpdateTraceId force updates the traced id to provided id
func UpdateTraceId(ctx context.Context, traceID string) context.Context {
	if traceID == "" {
		return SetTraceId(ctx)
	}
	ctx = loggers.AddToLogContext(ctx, "trace", traceID)
	return options.AddToOptions(ctx, tracerID, traceID)
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
