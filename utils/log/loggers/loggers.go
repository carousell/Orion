/*Package loggers provides loggers implementation for log package
 */
package loggers

import (
	"context"
	"fmt"
	"runtime"
	"strings"
)

type Label struct {
	Key   string      `json:"k"`
	Value interface{} `json:"v"`
}

const (
	// Key for log format version
	LogVersionKey = "_lv"
	// Log message format version
	LogVersion = "2"
	// Key for log message payload
	LogMessageKey = "msg"
	// Key for log labels
	LogLabelsKey = "lbl"
)

// Level type
type Level uint32

// Convert the Level to a string. E.g.  ErrorLevel becomes "error".
func (level Level) String() string {
	switch level {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case NoticeLevel:
		return "notice"
	case WarnLevel:
		return "warning"
	case ErrorLevel:
		return "error"
	case CriticalLevel:
		return "critical"
	}

	return "unknown"
}

// ParseLevel takes a string level and returns the log level constant.
func ParseLevel(lvl string) (Level, error) {
	switch strings.ToLower(lvl) {
	case "critical":
		return CriticalLevel, nil
	case "error":
		return ErrorLevel, nil
	case "warn", "warning":
		return WarnLevel, nil
	case "notice":
		return NoticeLevel, nil
	case "info":
		return InfoLevel, nil
	case "debug":
		return DebugLevel, nil
	}

	var l Level
	return l, fmt.Errorf("not a valid log Level: %q", lvl)
}

//AllLevels A constant exposing all logging levels
var AllLevels = []Level{
	CriticalLevel,
	ErrorLevel,
	NoticeLevel,
	WarnLevel,
	InfoLevel,
	DebugLevel,
}

// These are the different logging levels. You can set the logging level to log
// on your instance of logger, obtained with `logs.New()`.
const (
	// Critical level. Mostly for fatal error messages.
	CriticalLevel = iota
	// ErrorLevel level. Logs. Used for errors that should definitely be noted.
	// Commonly used for hooks to send errors to an error tracking service.
	ErrorLevel
	// WarnLevel level. Non-critical entries that deserve eyes.
	WarnLevel
	// Notice level. Should be used for logs that are not errors/warnings but are
	// necessary to be captured. This should not be as verbose as info.
	NoticeLevel
	// InfoLevel level. General operational entries about what's going on inside the
	// application.
	InfoLevel
	// DebugLevel level. Usually only enabled when debugging. Very verbose logging.
	DebugLevel
)

//BaseLogger is the interface that needs to be implemented by client loggers
type BaseLogger interface {
	Log(ctx context.Context, level Level, skip int, payload string, labels []Label, args ...interface{})
	SetLevel(level Level)
	GetLevel() Level
}

//Options contain all common options for BaseLoggers
type Options struct {
	ReplaceStdLogger   bool
	JSONLogs           bool
	Level              Level
	TimestampFieldName string
	LevelFieldName     string
	CallerInfo         bool
	CallerFileDepth    int
	CallerFieldName    string
}

//GetDefaultOptions fetches loggers default options
func GetDefaultOptions() Options {
	return DefaultOptions
}

// DefaultOptions stores all default options in loggers package
var (
	DefaultOptions = Options{
		ReplaceStdLogger:   false,
		JSONLogs:           true,
		Level:              NoticeLevel,
		TimestampFieldName: "@timestamp",
		LevelFieldName:     "level",
		CallerInfo:         true,
		CallerFileDepth:    2,
		CallerFieldName:    "caller",
	}
)

//Option defines an option for BaseLogger
type Option func(*Options)

//WithReplaceStdLogger enables/disables replacing std logger
func WithReplaceStdLogger(replaceStdLogger bool) Option {
	return func(o *Options) {
		o.ReplaceStdLogger = replaceStdLogger
	}
}

//WithJSONLogs enables/disables json logs
func WithJSONLogs(json bool) Option {
	return func(o *Options) {
		o.JSONLogs = json
	}
}

//WithTimestampFieldName sets the name of the time stamp field in logs
func WithTimestampFieldName(name string) Option {
	return func(o *Options) {
		if name != "" {
			o.TimestampFieldName = name
		}
	}
}

//WithLevelFieldName sets the name of the level field in logs
func WithLevelFieldName(name string) Option {
	return func(o *Options) {
		if name != "" {
			o.LevelFieldName = name
		}
	}
}

//WithCallerInfo enables/disables adding caller info to logs
func WithCallerInfo(callerInfo bool) Option {
	return func(o *Options) {
		o.CallerInfo = callerInfo
	}
}

//WithCallerFileDepth sets the depth of file to use in caller info
func WithCallerFileDepth(depth int) Option {
	return func(o *Options) {
		if depth > 0 {
			o.CallerFileDepth = depth
		}
	}
}

//WithCallerFieldName sets the name of callerinfo field
func WithCallerFieldName(name string) Option {
	return func(o *Options) {
		name = strings.TrimSpace(name)
		if name != "" {
			o.CallerFieldName = name
		}
	}
}

//FetchCallerInfo fetches function name, file name and line number from stack
func FetchCallerInfo(skip int, depth int) (function string, file string, line int) {
	if depth <= 0 {
		depth = 2
	}
	if skip < 0 {
		skip = 0
	}
	pc, file, line, _ := runtime.Caller(skip + 1)
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			depth--
			if depth == 0 {
				file = file[i+1:]
				break
			}
		}
	}
	return runtime.FuncForPC(pc).Name(), file, line
}
