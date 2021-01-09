package log

import (
	"github.com/carousell/Orion/utils/log/loggers"
	"github.com/carousell/log-svc-client"
)

type ConfigListener struct{}

func (l *ConfigListener) UpdateConfig(config log_svc_client.LogConfig) {
	logger := GetLogger()
	switch config.Level {
	case log_svc_client.CRITICAL:
		logger.SetLevel(loggers.ErrorLevel)
	case log_svc_client.ERROR:
		logger.SetLevel(loggers.ErrorLevel)
	case log_svc_client.WARNING:
		logger.SetLevel(loggers.WarnLevel)
	case log_svc_client.NOTICE:
		logger.SetLevel(loggers.InfoLevel)
	case log_svc_client.INFO:
		logger.SetLevel(loggers.InfoLevel)
	case log_svc_client.DEBUG:
		logger.SetLevel(loggers.DebugLevel)
	}
}
