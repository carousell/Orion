package notifier

func parseLevel(s string) severity {
	sev, ok := levelSeverityMap[s]
	if !ok {
		sev = errorSeverity // by default
	}

	return sev
}
