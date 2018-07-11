package loggers

import (
	"context"
	"fmt"
	"strings"
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
	case WarnLevel:
		return "warning"
	case ErrorLevel:
		return "error"
	}

	return "unknown"
}

// ParseLevel takes a string level and returns the log level constant.
func ParseLevel(lvl string) (Level, error) {
	switch strings.ToLower(lvl) {
	case "error":
		return ErrorLevel, nil
	case "warn", "warning":
		return WarnLevel, nil
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
	ErrorLevel,
	WarnLevel,
	InfoLevel,
	DebugLevel,
}

// These are the different logging levels. You can set the logging level to log
// on your instance of logger, obtained with `logs.New()`.
const (
	// ErrorLevel level. Logs. Used for errors that should definitely be noted.
	// Commonly used for hooks to send errors to an error tracking service.
	ErrorLevel = iota
	// WarnLevel level. Non-critical entries that deserve eyes.
	WarnLevel
	// InfoLevel level. General operational entries about what's going on inside the
	// application.
	InfoLevel
	// DebugLevel level. Usually only enabled when debugging. Very verbose logging.
	DebugLevel
)

//BaseLogger is the implementation that needs to be implemented by client loggers
type BaseLogger interface {
	Log(ctx context.Context, level Level, args ...interface{})
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
}

func GetDefaultOptions() Options {
	return DefaulOptions
}

var (
	DefaulOptions = Options{
		ReplaceStdLogger:   false,
		JSONLogs:           true,
		Level:              InfoLevel,
		TimestampFieldName: "@timestamp",
		LevelFieldName:     "level",
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

//WithReplaceStdLogger enables/disables replacing std logger
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
