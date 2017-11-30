package executor

import (
	"fmt"
	"sync"
)

type exe struct {
	wg      sync.WaitGroup
	work    chan Task
	config  config
	errc    chan error
	done    chan bool
	errored bool
}

type config struct {
	concurrency int
	failOnError bool
}

//WithConcurrency sets the number of concurrent works
func WithConcurrency(n int) Option {
	return func(c *config) {
		c.concurrency = n
	}
}

//WithFailOnError fails all task if even a single task returns a error
func WithFailOnError(fail bool) Option {
	return func(c *config) {
		c.failOnError = fail
	}
}

//Add adds a task to executor queue
func (e *exe) Add(task Task) {
	e.wg.Add(1)
	go e.add(task)
}

func (e *exe) add(task Task) {
	e.work <- task
}

func (e *exe) worker() {
	for t := range e.work {
		e.processTask(t)
	}
}

func (e *exe) processTask(task Task) {
	defer func(e *exe) {
		if r := recover(); r != nil {
			e.errc <- fmt.Errorf("PANIC: %s", r)
		}
		e.wg.Done()
	}(e)
	err := task()
	if err != nil {
		if e.config.failOnError {
			e.errc <- err
		}
	}
}

//Wait waits for all executors to finish or one of them to error based on option selected
func (e *exe) Wait() error {
	go func(e *exe) {
		e.wg.Wait()
		close(e.done)
		close(e.errc)
		close(e.work)
	}(e)
	for {
		select {
		case err := <-e.errc:
			if err != nil {
				e.errored = true
				go func(errc chan error) {
					for range errc {
					} // drain rest of the errors
				}(e.errc)
			}
			if e.config.failOnError {
				return err
			}
		case <-e.done:
			// rare corner case of errc trigring after done
			// when only 1 task was added
			select {
			case err := <-e.errc:
				return err
			default:
				return nil
			}
		}
	}
}

//NewExecutor builds and retuns a executor
func NewExecutor(options ...Option) Executor {
	c := config{
		concurrency: 5,
		failOnError: true,
	}
	for _, opt := range options {
		opt(&c)
	}
	e := new(exe)
	e.work = make(chan Task, 0)
	e.errc = make(chan error, 0)
	e.done = make(chan bool, 0)
	e.config = c
	for i := 0; i < c.concurrency; i++ {
		go e.worker()
	}
	return e
}
