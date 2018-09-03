package errors

const (
	// in the long term, we can define our own codes
	// now we're fully inheriting from grpc codes
	MaxCode = 20 // default value
)

type ReportLevel int

const (
	ReportLevelError ReportLevel = 1000
	ReportLevelWarn  ReportLevel = 1001
)
