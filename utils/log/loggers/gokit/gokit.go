package gokit

import (
	"context"
	stdlog "log"
	"os"

	"github.com/carousell/Orion/utils/log/loggers"
	"github.com/go-kit/kit/log"
)

type logger struct {
	logger log.Logger
	level  loggers.Level
}

func (l *logger) Log(ctx context.Context, level loggers.Level, args ...interface{}) {
	lgr := log.With(l.logger, "level", level.String())

	// fetch fields from context and add them to logrus fields
	ctxFields := loggers.FromContext(ctx)
	if ctxFields != nil {
		for k, v := range ctxFields {
			lgr = log.With(lgr, k, v)
		}
	}

	if len(args) == 1 {
		lgr.Log("msg", args[0])
	} else {
		lgr.Log(args...)
	}
}

func (l *logger) SetLevel(level loggers.Level) {
	l.level = level
}

func (l *logger) GetLevel() loggers.Level {
	return l.level
}

func NewLogger(options ...loggers.Option) loggers.BaseLogger {
	// default options
	opt := loggers.Options{
		ReplaceStdLogger: false,
	}
	// read options
	for _, f := range options {
		f(&opt)
	}

	l := logger{}
	l.logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout))
	l.level = loggers.InfoLevel
	if opt.ReplaceStdLogger {
		stdlog.SetOutput(log.NewStdlibAdapter(l.logger))
	}
	return &l
}
