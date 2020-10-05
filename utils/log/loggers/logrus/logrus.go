// Package logrus provides a BaseLogger implementation for logrus
package logrus

import (
	"context"
	"fmt"
	stdlog "log"
	"os"

	"github.com/carousell/Orion/utils/log/loggers"
	"github.com/sirupsen/logrus"
)

type logger struct {
	logger *logrus.Logger
	opt    loggers.Options
}

func toLogrusLogLevel(level loggers.Level) logrus.Level {
	switch level {
	case loggers.DebugLevel:
		return logrus.DebugLevel
	case loggers.InfoLevel:
		return logrus.InfoLevel
	case loggers.WarnLevel:
		return logrus.WarnLevel
	case loggers.ErrorLevel:
		return logrus.ErrorLevel
	case loggers.FatalLevel:
		return logrus.FatalLevel
	default:
		return logrus.ErrorLevel
	}
}

func (l *logger) Log(ctx context.Context, level loggers.Level, skip int, args ...interface{}) {
	fields := make(logrus.Fields)

	// fetch fields from context and add them to logrus fields
	ctxFields := loggers.FromContext(ctx)
	if ctxFields != nil {
		for k, v := range ctxFields {
			fields[k] = v
		}
	}

	if l.opt.CallerInfo {
		_, file, line := loggers.FetchCallerInfo(skip+1, l.opt.CallerFileDepth)
		fields[l.opt.CallerFieldName] = fmt.Sprintf("%s:%d", file, line)
	}

	logger := l.logger.WithFields(fields)
	switch level {
	case loggers.DebugLevel:
		logger.Debug(args...)
	case loggers.InfoLevel:
		logger.Info(args...)
	case loggers.WarnLevel:
		logger.Warn(args...)
	case loggers.ErrorLevel:
		logger.Error(args...)
	default:
		l.logger.Error(args...)
	}
}

func (l *logger) SetLevel(level loggers.Level) {
	l.logger.SetLevel(toLogrusLogLevel(level))
}

func (l *logger) GetLevel() loggers.Level {
	switch l.logger.Level {
	case logrus.DebugLevel:
		return loggers.DebugLevel
	case logrus.InfoLevel:
		return loggers.InfoLevel
	case logrus.WarnLevel:
		return loggers.WarnLevel
	case logrus.ErrorLevel:
		return loggers.ErrorLevel
	default:
		return loggers.InfoLevel
	}
}

//NewLogger returns a BaseLogger impl for logrus
func NewLogger(options ...loggers.Option) loggers.BaseLogger {
	// default options
	opt := loggers.GetDefaultOptions()
	// read options
	for _, f := range options {
		f(&opt)
	}

	l := logger{}
	l.logger = logrus.New()
	l.logger.Out = os.Stdout

	l.logger.SetLevel(toLogrusLogLevel(opt.Level))

	fieldMap := logrus.FieldMap{
		logrus.FieldKeyTime:  opt.TimestampFieldName,
		logrus.FieldKeyLevel: opt.LevelFieldName,
	}
	//check JSON logs
	if opt.JSONLogs {
		l.logger.Formatter = &logrus.JSONFormatter{
			FieldMap: fieldMap,
		}
	} else {
		l.logger.Formatter = &logrus.TextFormatter{
			FullTimestamp: true,
		}
	}

	l.opt = opt

	if opt.ReplaceStdLogger {
		stdlog.SetOutput(l.logger.Writer())
	}
	return &l
}
