package log

import (
	"context"

	"github.com/carousell/Orion/utils/log/loggers"
)

//SetLevel sets the log level to filter logs
func SetLevel(level loggers.Level) {
	GetLogger().SetLevel(level)
}

//GetLevel returns the current log level
func GetLevel() loggers.Level {
	return GetLogger().GetLevel()
}

//Debug writes out a debug log to global logger
func Debug(ctx context.Context, args ...interface{}) {
	GetLogger().Log(ctx, loggers.DebugLevel, 1, args...)
}

//Info writes out an info log to global logger
func Info(ctx context.Context, args ...interface{}) {
	GetLogger().Log(ctx, loggers.InfoLevel, 1, args...)
}

//Warn writes out a warning log to global logger
func Warn(ctx context.Context, args ...interface{}) {
	GetLogger().Log(ctx, loggers.WarnLevel, 1, args...)
}

//Error writes out an error log to global logger
func Error(ctx context.Context, args ...interface{}) {
	GetLogger().Log(ctx, loggers.ErrorLevel, 1, args...)
}
