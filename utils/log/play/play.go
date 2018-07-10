package main

import (
	"context"

	"github.com/carousell/Orion/utils/log"
	"github.com/carousell/Orion/utils/log/loggers"
)

func main() {
	ctx := context.Background()
	ctx = loggers.AddToLogContext(ctx, "hello", "world")
	log := log.GetLogger()
	//log := logs.NewLogger(logrus.NewLogger())
	log.SetLevel(loggers.InfoLevel)
	log.Error(ctx, "error")
	log.Warn(ctx, "warning")
	log.Info(ctx, "info")
	log.Debug(ctx, "debug")
	log.SetLevel(loggers.DebugLevel)
	log.Debug(ctx, "debug2")
}
