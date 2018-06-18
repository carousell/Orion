package worker

import (
	"context"
	"log"

	"github.com/RichardKnop/machinery/v1"
	machineryConfig "github.com/RichardKnop/machinery/v1/config"
	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/carousell/Orion/utils/spanutils"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

//NewWorker creates a new worker from given config
func NewWorker(config Config) Worker {
	w := new(worker)
	w.init(config)
	return w
}

type worker struct {
	server    *machinery.Server
	worker    *machinery.Worker
	config    Config
	LocalMap  map[string]wrappedWork
	localWork chan *workerInfo
}

//WithRetry sets the number of Retries for this task
func WithRetry(n int) ScheduleOption {
	return func(c *ScheduleConfig) {
		c.retries = n
	}
}

//WithQueueName sets the destination queue for this task
func WithQueueName(queueName string) ScheduleOption {
	return func(c *ScheduleConfig) {
		c.queueName = queueName
	}
}

func (w *worker) Schedule(ctx context.Context, name string, payload string, options ...ScheduleOption) error {
	span, ctx := spanutils.NewInternalSpan(ctx, name+"Scheduled")
	defer span.End()
	if w.config.LocalMode {
		return w.scheduleLocal(ctx, name, payload)
	}
	return w.scheduleRemote(ctx, name, payload, options...)
}

func (w *worker) scheduleRemote(ctx context.Context, name string, payload string, options ...ScheduleOption) error {
	c := ScheduleConfig{
		retries: 3,
	}
	for _, opt := range options {
		opt(&c)
	}
	wi := newWorkerInfo(ctx, payload)
	signature := &tasks.Signature{
		Name: name,
		Args: []tasks.Arg{
			{
				Type:  "string",
				Value: wi.String(),
			},
		},
	}
	signature.RetryCount = c.retries
	signature.RoutingKey = c.queueName
	_, err := w.server.SendTask(signature)
	if err != nil {
		return err
	}
	return nil

}

func (w *worker) scheduleLocal(ctx context.Context, name string, payload string) error {
	wi := newWorkerInfo(ctx, payload)
	wi.Name = name
	w.localWork <- wi
	return nil
}

func (w *worker) RegisterTask(name string, taskFunc Work) error {
	if w.config.LocalMode {
		w.LocalMap[name] = wrapperFunc(taskFunc)
	} else {
		return w.server.RegisterTask(name, wrapperFunc(taskFunc))
	}
	return nil
}

func wrapperFunc(task Work) wrappedWork {
	return func(payload string) error {
		wi := unmarshalWorkerInfo(payload)

		// rebuild span context
		wireContext, _ := opentracing.GlobalTracer().Extract(
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(wi.Trace))
		serverSpan := opentracing.StartSpan(
			wi.Name+"Worker",
			ext.RPCServerOption(wireContext))
		defer serverSpan.Finish()
		ctx := opentracing.ContextWithSpan(context.Background(), serverSpan)

		sp, ctx := spanutils.NewInternalSpan(ctx, wi.Name+"Process")
		defer sp.Finish()
		// execute task
		err := task(ctx, wi.Payload)
		if err != nil {
			sp.SetTag("error", err.Error())
		}
		return err
	}
}

func (w *worker) init(config Config) {
	w.config = buildConfig(config)
	if w.config.LocalMode {
		w.LocalMap = make(map[string]wrappedWork)
		w.localWork = make(chan *workerInfo, 100)
	} else if config.RabbitConfig != nil {
		//do machinery init
		rabbitServer := getServerName(config)
		var cfg = &machineryConfig.Config{
			Broker:       rabbitServer,
			DefaultQueue: config.RabbitConfig.QueueName,
			AMQP: &machineryConfig.AMQPConfig{
				Exchange:     "machinery_exchange",
				ExchangeType: "direct",
				BindingKey:   config.RabbitConfig.QueueName,
			},
			NoUnixSignals: true,
		}
		var err error
		w.server, err = machinery.NewServer(cfg)
		if err != nil {
			log.Println(err)
		}
		w.server.SetBackend(&fakeBackend{})
	}
}

func (w *worker) RunWorker(name string, concurrency int) {
	if w.config.LocalMode {
		for i := 0; i < concurrency; i++ {
			go func() {
				for wi := range w.localWork {
					if f, ok := w.LocalMap[wi.Name]; ok {
						f(wi.String())
					} else {
						log.Println("error", "could not find "+wi.Name)
					}
				}
			}()
		}
	} else {
		w.worker = w.server.NewWorker(name, concurrency)
		errc := make(chan error, 1)
		w.worker.LaunchAsync(errc)
	}
}

func (w *worker) CloseWorker() {
	if w.config.LocalMode {
		close(w.localWork)
	} else {
		w.worker.Quit()
	}
}
