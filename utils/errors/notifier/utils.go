package notifier

import (
	"github.com/carousell/Orion/utils/log/loggers"
	"github.com/getsentry/raven-go"
)

func (s severity) String() string {
	return string(s)
}

func (s severity) RavenSeverity() raven.Severity {
	return raven.Severity(s)
}

func (s severity) LoggerLevel() loggers.Level {
	switch s {
	case warningSeverity:
		return loggers.WarnLevel
	case infoSeverity:
		return loggers.InfoLevel
	case debugSeverity:
		return loggers.DebugLevel
	case errorSeverity:
		return loggers.ErrorLevel
	default:
		return loggers.ErrorLevel
	}
}

func parseLevel(s string) severity {
	sev, ok := levelSeverityMap[s]
	if !ok {
		sev = errorSeverity // by default
	}

	return sev
}
