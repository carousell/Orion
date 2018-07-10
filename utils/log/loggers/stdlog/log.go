package stdlog

import (
	"context"
	"log"

	"github.com/carousell/Orion/utils/log/loggers"
)

type logger struct {
	level loggers.Level
}

func (l *logger) SetLevel(level loggers.Level) {
	l.level = level
}

func (l *logger) GetLevel() loggers.Level {
	return l.level
}

func (l *logger) Log(ctx context.Context, level loggers.Level, args ...interface{}) {
	if l.level >= level {
		log.Println(args...)
	}
}

func NewLogger() loggers.BaseLogger {
	return &logger{
		level: loggers.InfoLevel,
	}
}
