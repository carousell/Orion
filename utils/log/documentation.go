/*Package log provides a minimal interface for structured logging in services.
Orion and orion utils use this log package for all logs.

How To Use

The simplest way to use this package is by calling static log functions to report particular level (critical/error/warning/notice/info/debug)
	log.Critical(...)
	log.Error(...)
	log.Warn(...)
	log.Notice(...)
	log.Info(...)
	log.Debug(...)

You can also initialize a new logger by calling 'log.NewLogger' and passing a loggers.BaseLogger implementation (loggers package provides a number of pre built implementations)
	logger := log.NewLogger(gokit.NewLogger())
	logger.Info(ctx, "log message", []loggers.Label{{"label_key", "label value"}})

Note:
	Preferred logging output is in either logfmt or json format, so to facilitate these log function arguments should be in pairs of key-value

Contextual Logs

log package uses context.Context to pass additional information to logs, you can use 'loggers.AddToLogContext' function to add additional information to logs. For example in access log from Orion service
	{"@timestamp":"2018-07-30T09:58:18.262948679Z","caller":"http/http.go:66","error":null,"grpcMethod":"/AuthSvc.AuthService/Authenticate","level":"info","method":"POST","path":"/2.0/authenticate/","took":"1.356812ms","trace":"15592e1b-93df-11e8-bdfd-0242ac110002","transport":"http"}
we pass 'method', 'path', 'grpcMethod' and 'transport' from context, this information gets automatically added to all log calls called inside the service and makes debugging services much easier.
Orion also generates a 'trace' ID per request, this can be used to trace an entire request path in logs.

Log Labels
log package provides ability to pass custom labels to logs. Labels can be used to search and filter specific logs. Label values can be any struct that can be marshalled as JSON.


*/
package log

//go:generate godoc2ghmd -ex -file=log/loggers/README.md github.com/carousell/Orion/utils/log/loggers
//go:generate godoc2ghmd -ex -file=log/loggers/gokit/README.md github.com/carousell/Orion/utils/log/loggers/gokit
//go:generate godoc2ghmd -ex -file=log/loggers/logrus/README.md github.com/carousell/Orion/utils/log/loggers/logrus
