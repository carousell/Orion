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
	log loggers.BaseLogger
}

func (l *logger) SetLevel(level loggers.Level) {
	l.log.SetLevel(level)
}

func (l *logger) GetLevel() loggers.Level {
	return l.log.GetLevel()
}

func (l *logger) Debug(ctx context.Context, args ...interface{}) {
	l.Log(ctx, loggers.DebugLevel, args...)
}

func (l *logger) Info(ctx context.Context, args ...interface{}) {
	l.Log(ctx, loggers.InfoLevel, args...)
}

func (l *logger) Warn(ctx context.Context, args ...interface{}) {
	l.Log(ctx, loggers.WarnLevel, args...)
}

func (l *logger) Error(ctx context.Context, args ...interface{}) {
	l.Log(ctx, loggers.ErrorLevel, args...)
}

func (l *logger) Log(ctx context.Context, level loggers.Level, args ...interface{}) {
	if ctx == nil {
		ctx = context.Background()
	}
	if l.GetLevel() >= level {
		l.log.Log(ctx, level, args...)
	}
}

//NewLogger creates a new logger
func NewLogger(log loggers.BaseLogger) Logger {
	l := new(logger)
	l.log = log
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
