# worker
`import "github.com/carousell/Orion/utils/worker"`

* [Overview](#pkg-overview)
* [Imported Packages](#pkg-imports)
* [Index](#pkg-index)

## <a name="pkg-overview">Overview</a>

## <a name="pkg-imports">Imported Packages</a>

- [github.com/carousell/go-utils/utils/spanutils](./../../../go-utils/utils/spanutils)
- [github.com/carousell/machinery/v1](./../../../machinery/v1)
- [github.com/carousell/machinery/v1/config](./../../../machinery/v1/config)
- [github.com/carousell/machinery/v1/tasks](./../../../machinery/v1/tasks)
- [github.com/grpc-ecosystem/go-grpc-middleware/util/metautils](https://godoc.org/github.com/grpc-ecosystem/go-grpc-middleware/util/metautils)
- [github.com/opentracing/opentracing-go](https://godoc.org/github.com/opentracing/opentracing-go)
- [github.com/opentracing/opentracing-go/ext](https://godoc.org/github.com/opentracing/opentracing-go/ext)
- [github.com/satori/go.uuid](https://godoc.org/github.com/satori/go.uuid)

## <a name="pkg-index">Index</a>
* [type Config](#Config)
* [type RabbitMQConfig](#RabbitMQConfig)
* [type ScheduleConfig](#ScheduleConfig)
* [type ScheduleOption](#ScheduleOption)
  * [func WithRetry(n int) ScheduleOption](#WithRetry)
* [type Work](#Work)
* [type Worker](#Worker)
  * [func NewWorker(config Config) Worker](#NewWorker)

#### <a name="pkg-files">Package files</a>
[config.go](./config.go) [types.go](./types.go) [worker.go](./worker.go) [workerinfo.go](./workerinfo.go) 

## <a name="Config">type</a> [Config](./types.go#L18-L21)
``` go
type Config struct {
    LocalMode    bool
    RabbitConfig *RabbitMQConfig
}
```
Config is the config used to intialize workers

## <a name="RabbitMQConfig">type</a> [RabbitMQConfig](./types.go#L24-L31)
``` go
type RabbitMQConfig struct {
    UserName    string
    Password    string
    BrokerVHost string
    Host        string
    Port        string
    QueueName   string
}
```
RabbitMQConfig is the config used for scheduling tasks through rabbitmq

## <a name="ScheduleConfig">type</a> [ScheduleConfig](./types.go#L38-L40)
``` go
type ScheduleConfig struct {
    // contains filtered or unexported fields
}
```
ScheduleConfig is the config used when scheduling a task

## <a name="ScheduleOption">type</a> [ScheduleOption](./types.go#L43)
``` go
type ScheduleOption func(*ScheduleConfig)
```
ScheduleOption represents different options available for Schedule

### <a name="WithRetry">func</a> [WithRetry](./worker.go#L31)
``` go
func WithRetry(n int) ScheduleOption
```
WithRetry sets the number of Retries for this task

## <a name="Work">type</a> [Work](./types.go#L34)
``` go
type Work func(ctx context.Context, payload string) error
```
Work is the type of task that can be exeucted by Worker

## <a name="Worker">type</a> [Worker](./types.go#L10-L15)
``` go
type Worker interface {
    Schedule(ctx context.Context, name, payload string, options ...ScheduleOption) error
    RegisterTask(name string, taskFunc Work) error
    RunWorker(name string, concurrency int)
    CloseWorker()
}
```
Worker is the interface for worker

### <a name="NewWorker">func</a> [NewWorker](./worker.go#L16)
``` go
func NewWorker(config Config) Worker
```
NewWorker creates a new worker from given config

- - -
Generated by [godoc2ghmd](https://github.com/GandalfUK/godoc2ghmd)