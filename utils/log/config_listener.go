package log

import (
	"context"
	"github.com/carousell/Orion/utils/log/loggers"
	"github.com/carousell/log-svc-client"
	log "github.com/sirupsen/logrus"
)

type ConfigListener struct{}

func (l *ConfigListener) UpdateConfig(config log_svc_client.LogConfig) {
	log.Info(context.Background(), "Updating log level")
	logger := GetLogger()
	switch config.Level {
	case log_svc_client.CRITICAL:
		log.Info("Setting log level to ERROR")
		logger.SetLevel(loggers.ErrorLevel)
	case log_svc_client.ERROR:
		log.Info("Setting log level to ERROR")
		logger.SetLevel(loggers.ErrorLevel)
	case log_svc_client.WARNING:
		log.Info("Setting log level to WARNING")
		logger.SetLevel(loggers.WarnLevel)
	case log_svc_client.NOTICE:
		log.Info("Setting log level to INFO")
		logger.SetLevel(loggers.InfoLevel)
	case log_svc_client.INFO:
		log.Info("Setting log level to INFO")
		logger.SetLevel(loggers.InfoLevel)
	case log_svc_client.DEBUG:
		log.Info("Setting log level to DEBUG")
		logger.SetLevel(loggers.DebugLevel)
	}
}
