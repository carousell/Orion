package main

import (
	"context"
	"fmt"
	"github.com/carousell/Orion/utils/errors"
	"github.com/carousell/Orion/utils/errors/notifier"
	"github.com/carousell/Orion/utils/log"
	"github.com/carousell/Orion/utils/log/loggers"
)

func main() {
	ctx := context.Background()
	ctx = loggers.AddToLogContext(ctx, "hello", "world")
	logger := log.GetLogger()
	//logger := log.NewLogger(gokit.NewLogger())
	logger.SetLevel(loggers.InfoLevel)
	logger.Error(ctx, "error", nil)
	logger.Warn(ctx, "warning", nil)
	logger.Info(ctx, "info", []loggers.Label{{"hello", "hi"}})
	logger.Debug(ctx, "debug", nil)
	logger.SetLevel(loggers.DebugLevel)
	logger.Debug(ctx, "debug2", nil)
	log.Debug(ctx, "debug3", nil)

	e := errors.New("hello world")
	notifier.Notify(e, ctx)

	e2 := fmt.Errorf("generic error")
	notifier.Notify(e2, ctx)
}
