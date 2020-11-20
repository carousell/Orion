package log

import (
	"context"

	"github.com/carousell/Orion/utils/log/loggers"
)

// Logger interface is implemnted by the log implementation
type Logger interface {
	loggers.BaseLogger
	Debug(ctx context.Context, payload string, labels []loggers.Label, args ...interface{})
	Info(ctx context.Context, payload string, labels []loggers.Label, args ...interface{})
	Notice(ctx context.Context, payload string, labels []loggers.Label, args ...interface{})
	Warn(ctx context.Context, payload string, labels []loggers.Label, args ...interface{})
	Error(ctx context.Context, payload string, labels []loggers.Label, args ...interface{})
	Critical(ctx context.Context, payload string, labels []loggers.Label, args ...interface{})
}
