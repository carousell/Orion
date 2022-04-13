package notifier

import (
	"context"
	bugsnag "github.com/bugsnag/bugsnag-go"
	"github.com/carousell/Orion/utils/errors"
	"os"
	"reflect"
	"runtime"
	"strconv"

	"github.com/carousell/Orion/utils/log"
	"github.com/carousell/Orion/utils/log/loggers"
	raven "github.com/getsentry/raven-go"
	stdopentracing "github.com/opentracing/opentracing-go"
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

type isTags interface {
	isTags()
	value() map[string]string
}

type Tags map[string]string

func (tags Tags) isTags() {}

func (tags Tags) value() map[string]string {
	return map[string]string(tags)
}

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

func convToSentry(in errors.ErrorExt) *raven.Stacktrace {
	out := new(raven.Stacktrace)
	pcs := in.Callers()
	frames := make([]*raven.StacktraceFrame, 0)

	callersFrames := runtime.CallersFrames(pcs)

	for {
		fr, more := callersFrames.Next()
		if fr.Func != nil {
			frame := raven.NewStacktraceFrame(fr.PC, fr.Function, fr.File, fr.Line, 3, []string{})
			if frame != nil {
				frame.InApp = true
				frames = append(frames, frame)
			}
		}
		if !more {
			break
		}
	}
	for i := len(frames)/2 - 1; i >= 0; i-- {
		opp := len(frames) - 1 - i
		frames[i], frames[opp] = frames[opp], frames[i]
	}
	out.Frames = frames
	return out
}

func parseRawData(ctx context.Context, rawData ...interface{}) (extraData map[string]interface{}, tagData []map[string]string) {
	extraData = make(map[string]interface{})

	for pos := range rawData {
		data := rawData[pos]
		if _, ok := data.(context.Context); ok {
			continue
		}
		if tags, ok := data.(isTags); ok {
			tagData = append(tagData, tags.value())
		} else {
			extraData[reflect.TypeOf(data).String()+strconv.Itoa(pos)] = data
		}
	}
	if logFields := loggers.FromContext(ctx); logFields != nil {
		for k, v := range logFields {
			extraData[k] = v
		}
	}
	return
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
	sev := parseLevel(level)

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
			ctx = c
			break
		}
	}

	if airbrake != nil {
		var n *gobrake.Notice
		n = gobrake.NewNotice(errWithStack, nil, 1)
		n.Errors[0].Backtrace = convToGoBrake(errWithStack.StackFrame())
		if len(list) > 0 {
			m, _ := parseRawData(ctx, list...)
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
	parsedData, tagData := parseRawData(ctx, list...)
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
		rollbar.ErrorWithStack(sev.String(), errWithStack, convToRollbar(errWithStack.StackFrame()), fields...)
	}

	if sentryInited {
		ravenExp := raven.NewException(errWithStack, convToSentry(errWithStack))
		packet := raven.NewPacketWithExtra(errWithStack.Error(), parsedData, ravenExp)

		for _, tags := range tagData {
			packet.AddTags(tags)
		}

		// type assert directly since it's single use case so we don't consider about wrapping it for now
		packet.Level = raven.Severity(sev)
		raven.Capture(packet, nil)
	}

	log.GetLogger().Log(ctx, sev.LoggerLevel(), skip+1, "err", errWithStack, "stack", errWithStack.StackFrame())
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
		parsedData, tagData := parseRawData(ctx, rawData...)
		if rollbarInited {
			rollbar.ErrorWithStack(rollbar.CRIT, e, convToRollbar(e.StackFrame()), &rollbar.Field{Name: "panic", Data: r})
		}
		if sentryInited {
			ravenExp := raven.NewException(e, convToSentry(e))
			packet := raven.NewPacketWithExtra(e.Error(), parsedData, ravenExp)

			for _, tags := range tagData {
				packet.AddTags(tags)
			}

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
