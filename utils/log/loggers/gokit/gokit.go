//Package gokit provides BaseLogger implementation for go-kit/log
package gokit

import (
	"context"
	"fmt"
	stdlog "log"
	"os"

	"github.com/carousell/Orion/utils/log/loggers"
	"github.com/go-kit/kit/log"
)

type logger struct {
	logger log.Logger
	level  loggers.Level
	opt    loggers.Options
}

func (l *logger) Log(ctx context.Context, level loggers.Level, skip int, args ...interface{}) {
	lgr := log.With(l.logger, l.opt.LevelFieldName, level.String())

	if l.opt.CallerInfo {
		_, file, line := loggers.FetchCallerInfo(skip+1, l.opt.CallerFileDepth)
		lgr = log.With(lgr, l.opt.CallerFieldName, fmt.Sprintf("%s:%d", file, line))
	}

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

//NewLogger returns a base logger impl for go-kit log
func NewLogger(options ...loggers.Option) loggers.BaseLogger {
	// default options
	opt := loggers.GetDefaultOptions()

	// read options
	for _, f := range options {
		f(&opt)
	}

	l := logger{}
	writer := log.NewSyncWriter(os.Stdout)

	// check for json or logfmt
	if opt.JSONLogs {
		l.logger = log.NewJSONLogger(writer)
	} else {
		l.logger = log.NewLogfmtLogger(writer)
	}

	l.logger = log.With(l.logger, opt.TimestampFieldName, log.DefaultTimestamp)

	l.level = opt.Level
	l.opt = opt

	if opt.ReplaceStdLogger {
		stdlog.SetFlags(stdlog.LUTC)
		stdlog.SetOutput(log.NewStdlibAdapter(l.logger, log.TimestampKey(opt.TimestampFieldName)))
	}
	return &l
}
