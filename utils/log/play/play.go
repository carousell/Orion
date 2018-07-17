package main

import (
	"context"

	"github.com/carousell/Orion/utils/log"
	"github.com/carousell/Orion/utils/log/loggers"
)

func main() {
	ctx := context.Background()
	ctx = loggers.AddToLogContext(ctx, "hello", "world")
	logger := log.GetLogger()
	//logger := log.NewLogger(stdlog.NewLogger())
	logger.SetLevel(loggers.InfoLevel)
	logger.Error(ctx, "error")
	logger.Warn(ctx, "warning")
	logger.Info(ctx, "info")
	logger.Debug(ctx, "debug")
	logger.SetLevel(loggers.DebugLevel)
	logger.Debug(ctx, "debug2")
	log.Debug(ctx, "debug3")
}
