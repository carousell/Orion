package notifier

import (
	"context"
	"fmt"
	"go/build"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	bugsnag "github.com/bugsnag/bugsnag-go"
	"github.com/carousell/Orion/utils/errors"
	"github.com/carousell/Orion/utils/log"
	"github.com/carousell/Orion/utils/log/loggers"
	"github.com/carousell/Orion/utils/options"
	sentry "github.com/getsentry/sentry-go"
	stdopentracing "github.com/opentracing/opentracing-go"
	"github.com/pborman/uuid"
	"github.com/stvp/rollbar"
	gobrake "gopkg.in/airbrake/gobrake.v2"
)

const (
	tracerID = "tracerId"
)

var (
	airbrake      *gobrake.Notifier
	bugsnagInited bool
	rollbarInited bool
	sentryInited  bool
	serverRoot    string
	hostname      string
	// Used for populating a list of source directories
	// to be compared against for trimming of filenames
	trimPaths []string
)

func init() {
	// Collect all source directories, and make sure they
	// end in a trailing "separator"
	for _, prefix := range build.Default.SrcDirs() {
		if prefix[len(prefix)-1] != filepath.Separator {
			prefix += string(filepath.Separator)
		}
		trimPaths = append(trimPaths, prefix)
	}
}

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

// InitSentry inits sentry configuration
func InitSentry(dsn, env string) {
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              dsn,
		Debug:            true,
		AttachStacktrace: true,
		Environment:      env,
		ServerName:       getHostname(),
		MaxBreadcrumbs:   30,
		// Remove modules integration from being inited
		// Add more integrations here if required
		Integrations: func(integrations []sentry.Integration) []sentry.Integration {
			var filteredIntegrations []sentry.Integration
			for _, integration := range integrations {
				if integration.Name() == "Modules" {
					continue
				}
				filteredIntegrations = append(filteredIntegrations, integration)
			}
			return filteredIntegrations
		},
	})

	sentryInited = true
	if err != nil {
		fmt.Printf("Sentry initialization failed: %v\n", err)
		sentryInited = false
	}
}

func Notify(err error, rawData ...interface{}) error {
	return NotifyWithLevelAndSkip(err, 2, loggers.Level(loggers.ErrorLevel).String(), rawData...)
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

	parsedData, tagData := parseRawData(ctx, list...)

	if airbrake != nil {
		doNotifyAirbrake(ctx, errWithStack, traceID, list...)
	}
	if bugsnagInited {
		doNotifyBugsnag(errWithStack, list...)
	}
	if rollbarInited {
		doNotifyRollbar(errWithStack, level, parsedData, traceID, list...)
	}
	if sentryInited {
		doNotifySentry(errWithStack, level, parsedData, tagData)
	}

	log.GetLogger().Log(ctx, loggers.ErrorLevel, skip+1, "err", errWithStack, "stack", errWithStack.StackFrame())
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
			doNotifySentry(e, string(sentry.LevelFatal), parsedData, tagData)
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
