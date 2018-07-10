package log

import (
	"context"

	"github.com/carousell/Orion/utils/log/loggers"
)

func SetLevel(level loggers.Level) {
	GetLogger().SetLevel(level)
}

func GetLevel() loggers.Level {
	return GetLogger().GetLevel()
}

func Debug(ctx context.Context, args ...interface{}) {
	GetLogger().Debug(ctx, args...)
}

func Info(ctx context.Context, args ...interface{}) {
	GetLogger().Info(ctx, args...)
}

func Warn(ctx context.Context, args ...interface{}) {
	GetLogger().Warn(ctx, args...)
}

func Error(ctx context.Context, args ...interface{}) {
	GetLogger().Error(ctx, args...)
}
