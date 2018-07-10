package log

import (
	"context"

	"github.com/carousell/Orion/utils/log/loggers"
)

// Logger interface is implemnted by the log implementation
type Logger interface {
	SetLevel(level loggers.Level)
	GetLevel() loggers.Level
	Debug(ctx context.Context, args ...interface{})
	Info(ctx context.Context, args ...interface{})
	Warn(ctx context.Context, args ...interface{})
	Error(ctx context.Context, args ...interface{})
}
