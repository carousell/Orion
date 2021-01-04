package notifier

import (
	"context"
	"os"
	"reflect"
	"runtime"
	"strings"

	bugsnag "github.com/bugsnag/bugsnag-go"
	"github.com/carousell/Orion/utils/errors"
	sentry "github.com/getsentry/sentry-go"
	"github.com/stvp/rollbar"
	gobrake "gopkg.in/airbrake/gobrake.v2"
)

func doNotifyAirbrake(ctx context.Context, errWithStack errors.ErrorExt, traceID string, list ...interface{}) {
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

func doNotifyBugsnag(errWithStack errors.ErrorExt, list ...interface{}) {
	bugsnag.Notify(errWithStack, list...)
}

func doNotifyRollbar(errWithStack errors.ErrorExt, level string, parsedData map[string]interface{}, traceID string, list ...interface{}) {
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

func doNotifySentry(errWithStack errors.ErrorExt, level string, parsedData map[string]interface{}, tagData []map[string]string) {
	sentryLevel := sentry.LevelError
	if level == "critical" {
		sentryLevel = sentry.LevelFatal
	} else if level == "warning" {
		sentryLevel = sentry.LevelWarning
	}

	// Use copy of hub for deterministic and safe
	// calls to the Sentry Hub across goroutines
	// ref: https://docs.sentry.io/platforms/go/goroutines/
	hub := sentry.CurrentHub().Clone()

	event := sentry.NewEvent()
	event.Message = errWithStack.Error()
	event.Exception = []sentry.Exception{
		{
			Value:      errWithStack.Error(),
			Type:       reflect.TypeOf(errWithStack).String(),
			Stacktrace: convToSentry(errWithStack),
		},
	}
	hub.ConfigureScope(func(scope *sentry.Scope) {
		// Extract user and request from parsed data to
		// explicitly set them so that it can be filtered
		// upon in the dashboard
		if u := extractSentryUserFromParsedData(parsedData); u != nil {
			scope.SetUser(*u)
		}

		for _, tags := range tagData {
			scope.SetTags(tags)
		}
		scope.SetExtras(parsedData)
		scope.SetLevel(sentryLevel)

	})
	hub.CaptureEvent(event)
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

func convToSentry(in errors.ErrorExt) *sentry.Stacktrace {
	out := sentry.NewStacktrace()
	pcs := in.Callers()
	frames := make([]sentry.Frame, 0)

	callersFrames := runtime.CallersFrames(pcs)

	for {
		fr, more := callersFrames.Next()
		if fr.Func != nil {
			module, function := deconstructFunctionName(fr.Function)
			frame := sentry.Frame{
				Filename: trimPath(fr.File),
				AbsPath:  trimPath(fr.File),
				Lineno:   fr.Line,
				Function: function,
				Module:   module,
				InApp:    true,
			}
			frames = append(frames, frame)
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

// Try to trim the GOROOT or GOPATH prefix off of a filename
func trimPath(filename string) string {
	for _, prefix := range trimPaths {
		if trimmed := strings.TrimPrefix(filename, prefix); len(trimmed) < len(filename) {
			return trimmed
		}
	}
	return filename
}

// Transform `runtime/debug.*T·ptrmethod` into `{ module: runtime/debug, function: *T.ptrmethod }`
func deconstructFunctionName(name string) (module string, function string) {
	if idx := strings.LastIndex(name, "."); idx != -1 {
		module = name[:idx]
		function = name[idx+1:]
	}
	function = strings.Replace(function, "·", ".", -1)
	return module, function
}

// extractUserFromParsedData returns the first sentry.User object found in the parsedData
func extractSentryUserFromParsedData(parsedData map[string]interface{}) *sentry.User {
	for pos := range parsedData {
		data := parsedData[pos]
		if u, ok := data.(sentry.User); ok {
			return &u
		}
	}
	return nil
}
