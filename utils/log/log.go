package log

import (
	"context"
	"sync"

	"github.com/carousell/Orion/utils/log/loggers"
	"github.com/carousell/Orion/utils/log/loggers/gokit"
)

var (
	defaultLogger Logger
	mu            sync.Mutex
	once          sync.Once
)

type logger struct {
	baseLog loggers.BaseLogger
}

func (l *logger) SetLevel(level loggers.Level) {
	l.baseLog.SetLevel(level)
}

func (l *logger) GetLevel() loggers.Level {
	return l.baseLog.GetLevel()
}

func (l *logger) Debug(ctx context.Context, payload string, labels []loggers.Label, args ...interface{}) {
	l.Log(ctx, loggers.DebugLevel, 1, payload, labels, args...)
}

func (l *logger) Info(ctx context.Context, payload string, labels []loggers.Label, args ...interface{}) {
	l.Log(ctx, loggers.InfoLevel, 1, payload, labels, args...)
}

func (l *logger) Notice(ctx context.Context, payload string, labels []loggers.Label, args ...interface{}) {
	l.Log(ctx, loggers.NoticeLevel, 1, payload, labels, args...)
}

func (l *logger) Warn(ctx context.Context, payload string, labels []loggers.Label, args ...interface{}) {
	l.Log(ctx, loggers.WarnLevel, 1, payload, labels, args...)
}

func (l *logger) Error(ctx context.Context, payload string, labels []loggers.Label, args ...interface{}) {
	l.Log(ctx, loggers.ErrorLevel, 1, payload, labels, args...)
}

func (l *logger) Critical(ctx context.Context, payload string, labels []loggers.Label, args ...interface{}) {
	l.Log(ctx, loggers.CriticalLevel, 1, payload, labels, args...)
}

func (l *logger) Log(ctx context.Context, level loggers.Level, skip int, payload string, labels []loggers.Label, args ...interface{}) {
	if ctx == nil {
		ctx = context.Background()
	}
	if l.GetLevel() >= level {
		l.baseLog.Log(ctx, level, skip+1, payload, labels, args...)
	}
}

//NewLogger creates a new logger with a provided BaseLogger
func NewLogger(log loggers.BaseLogger) Logger {
	l := new(logger)
	l.baseLog = log
	return l
}

//GetLogger returns the global logger
func GetLogger() Logger {
	if defaultLogger == nil {
		once.Do(func() {
			defaultLogger = NewLogger(gokit.NewLogger())
		})
	}
	return defaultLogger
}

//SetLogger sets the global logger
func SetLogger(l Logger) {
	if l != nil {
		mu.Lock()
		defer mu.Unlock()
		defaultLogger = l
	}
}
