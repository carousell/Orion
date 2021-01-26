//Package stdlog provides a BaseLogger implementation for golang "log" package
package stdlog

import (
	"context"
	"log"

	"github.com/carousell/Orion/utils/log/loggers"
)

type logger struct {
	level        loggers.Level
	samplingRate int
}

func (l *logger) SetLevel(level loggers.Level) {
	l.level = level
}

func (l *logger) GetLevel() loggers.Level {
	return l.level
}

func (l *logger) SetSampling(percent int) {
	l.samplingRate = percent
}

func (l *logger) GetSampling() int {
	return l.samplingRate
}

func (l *logger) Log(ctx context.Context, level loggers.Level, skip int, args ...interface{}) {
	if l.level >= level {
		// fetch fields from context and add them to logrus fields
		ctxFields := loggers.FromContext(ctx)
		if ctxFields != nil {
			for k, v := range ctxFields {
				args = append(args, k, v)
			}
		}
		log.Println(args...)
	}
}

//NewLogger returns a BaseLogger impl for golang "log" package
func NewLogger(options ...loggers.Option) loggers.BaseLogger {
	opt := loggers.GetDefaultOptions()

	return &logger{
		level:        opt.Level,
		samplingRate: opt.SamplingRate,
	}
}
