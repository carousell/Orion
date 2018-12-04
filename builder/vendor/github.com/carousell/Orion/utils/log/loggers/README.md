# loggers
`import "github.com/carousell/Orion/utils/log/loggers"`

* [Overview](#pkg-overview)
* [Imported Packages](#pkg-imports)
* [Index](#pkg-index)

## <a name="pkg-overview">Overview</a>
Package loggers provides loggers implementation for log package

## <a name="pkg-imports">Imported Packages</a>

No packages beyond the Go standard library are imported.

## <a name="pkg-index">Index</a>
* [Constants](#pkg-constants)
* [Variables](#pkg-variables)
* [func AddToLogContext(ctx context.Context, key string, value interface{}) context.Context](#AddToLogContext)
* [func FetchCallerInfo(skip int, depth int) (function string, file string, line int)](#FetchCallerInfo)
* [type BaseLogger](#BaseLogger)
* [type Level](#Level)
  * [func ParseLevel(lvl string) (Level, error)](#ParseLevel)
  * [func (level Level) String() string](#Level.String)
* [type LogFields](#LogFields)
  * [func FromContext(ctx context.Context) LogFields](#FromContext)
  * [func (o LogFields) Add(key string, value interface{})](#LogFields.Add)
  * [func (o LogFields) Del(key string)](#LogFields.Del)
* [type Option](#Option)
  * [func WithCallerFileDepth(depth int) Option](#WithCallerFileDepth)
  * [func WithCallerInfo(callerInfo bool) Option](#WithCallerInfo)
  * [func WithJSONLogs(json bool) Option](#WithJSONLogs)
  * [func WithLevelFieldName(name string) Option](#WithLevelFieldName)
  * [func WithReplaceStdLogger(replaceStdLogger bool) Option](#WithReplaceStdLogger)
  * [func WithTimestampFieldName(name string) Option](#WithTimestampFieldName)
* [type Options](#Options)
  * [func GetDefaultOptions() Options](#GetDefaultOptions)

#### <a name="pkg-files">Package files</a>
[fields.go](./fields.go) [loggers.go](./loggers.go) 

## <a name="pkg-constants">Constants</a>
``` go
const (
    // ErrorLevel level. Logs. Used for errors that should definitely be noted.
    // Commonly used for hooks to send errors to an error tracking service.
    ErrorLevel = iota
    // WarnLevel level. Non-critical entries that deserve eyes.
    WarnLevel
    // InfoLevel level. General operational entries about what's going on inside the
    // application.
    InfoLevel
    // DebugLevel level. Usually only enabled when debugging. Very verbose logging.
    DebugLevel
)
```
These are the different logging levels. You can set the logging level to log
on your instance of logger, obtained with `logs.New()`.

## <a name="pkg-variables">Variables</a>
``` go
var AllLevels = []Level{
    ErrorLevel,
    WarnLevel,
    InfoLevel,
    DebugLevel,
}
```
AllLevels A constant exposing all logging levels

``` go
var (
    DefaultOptions = Options{
        ReplaceStdLogger:   false,
        JSONLogs:           true,
        Level:              InfoLevel,
        TimestampFieldName: "@timestamp",
        LevelFieldName:     "level",
        CallerInfo:         true,
        CallerFileDepth:    2,
    }
)
```
DefaultOptions stores all default options in loggers package

## <a name="AddToLogContext">func</a> [AddToLogContext](./fields.go#L30)
``` go
func AddToLogContext(ctx context.Context, key string, value interface{}) context.Context
```
AddToLogContext adds log fields to context.
Any info added here will be added to all logs using this context

## <a name="FetchCallerInfo">func</a> [FetchCallerInfo](./loggers.go#L159)
``` go
func FetchCallerInfo(skip int, depth int) (function string, file string, line int)
```
FetchCallerInfo fetches function name, file name and line number from stack

## <a name="BaseLogger">type</a> [BaseLogger](./loggers.go#L72-L76)
``` go
type BaseLogger interface {
    Log(ctx context.Context, level Level, skip int, args ...interface{})
    SetLevel(level Level)
    GetLevel() Level
}
```
BaseLogger is the interface that needs to be implemented by client loggers

## <a name="Level">type</a> [Level](./loggers.go#L13)
``` go
type Level uint32
```
Level type

### <a name="ParseLevel">func</a> [ParseLevel](./loggers.go#L32)
``` go
func ParseLevel(lvl string) (Level, error)
```
ParseLevel takes a string level and returns the log level constant.

### <a name="Level.String">func</a> (Level) [String](./loggers.go#L16)
``` go
func (level Level) String() string
```
Convert the Level to a string. E.g.  ErrorLevel becomes "error".

## <a name="LogFields">type</a> [LogFields](./fields.go#L14)
``` go
type LogFields map[string]interface{}
```
LogFields contains all fields that have to be added to logs

### <a name="FromContext">func</a> [FromContext](./fields.go#L44)
``` go
func FromContext(ctx context.Context) LogFields
```
FromContext fetchs log fields from provided context

### <a name="LogFields.Add">func</a> (LogFields) [Add](./fields.go#L17)
``` go
func (o LogFields) Add(key string, value interface{})
```
Add or modify log fields

### <a name="LogFields.Del">func</a> (LogFields) [Del](./fields.go#L24)
``` go
func (o LogFields) Del(key string)
```
Del deletes a log field entry

## <a name="Option">type</a> [Option](./loggers.go#L108)
``` go
type Option func(*Options)
```
Option defines an option for BaseLogger

### <a name="WithCallerFileDepth">func</a> [WithCallerFileDepth](./loggers.go#L150)
``` go
func WithCallerFileDepth(depth int) Option
```
WithCallerFileDepth sets the depth of file to use in caller info

### <a name="WithCallerInfo">func</a> [WithCallerInfo](./loggers.go#L143)
``` go
func WithCallerInfo(callerInfo bool) Option
```
WithCallerInfo enables/disables adding caller info to logs

### <a name="WithJSONLogs">func</a> [WithJSONLogs](./loggers.go#L118)
``` go
func WithJSONLogs(json bool) Option
```
WithJSONLogs enables/disables json logs

### <a name="WithLevelFieldName">func</a> [WithLevelFieldName](./loggers.go#L134)
``` go
func WithLevelFieldName(name string) Option
```
WithLevelFieldName sets the name of the level field in logs

### <a name="WithReplaceStdLogger">func</a> [WithReplaceStdLogger](./loggers.go#L111)
``` go
func WithReplaceStdLogger(replaceStdLogger bool) Option
```
WithReplaceStdLogger enables/disables replacing std logger

### <a name="WithTimestampFieldName">func</a> [WithTimestampFieldName](./loggers.go#L125)
``` go
func WithTimestampFieldName(name string) Option
```
WithTimestampFieldName sets the name of the time stamp field in logs

## <a name="Options">type</a> [Options](./loggers.go#L79-L87)
``` go
type Options struct {
    ReplaceStdLogger   bool
    JSONLogs           bool
    Level              Level
    TimestampFieldName string
    LevelFieldName     string
    CallerInfo         bool
    CallerFileDepth    int
}
```
Options contain all common options for BaseLoggers

### <a name="GetDefaultOptions">func</a> [GetDefaultOptions](./loggers.go#L90)
``` go
func GetDefaultOptions() Options
```
GetDefaultOptions fetches loggers default options

- - -
Generated by [godoc2ghmd](https://github.com/GandalfUK/godoc2ghmd)