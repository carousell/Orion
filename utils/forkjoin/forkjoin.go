package forkjoin

import (
	"fmt"
	"sync"
)

func NewForkJoin() ForkJoin {
	return &forkJoin{
		errc: make(chan error, 0),
		done: make(chan bool, 0),
	}
}

type forkJoin struct {
	wg      sync.WaitGroup
	errc    chan error
	done    chan bool
	errored bool
}

func (f *forkJoin) Add(task Task) {
	// if we have already errored dont start task
	if !f.errored {
		f.wg.Add(1)
		go f.startTask(task)
	}
}

func (f *forkJoin) startTask(task Task) {
	defer func(f *forkJoin) {
		if r := recover(); r != nil {
			f.errc <- fmt.Errorf("PANIC: %s", r)
		}
		f.wg.Done()
	}(f)
	err := task()
	if err != nil {
		f.errc <- err
	}
}

func (f *forkJoin) Wait() error {
	go func(f *forkJoin) {
		f.wg.Wait()
		close(f.done)
		close(f.errc)
	}(f)
	select {
	case err := <-f.errc:
		if err != nil {
			f.errored = true
			go func(errc chan error) {
				for range errc {
				} // drain rest of the errors
			}(f.errc)
		}
		return err
	case <-f.done:
		// rare corner case of errc trigring after done
		// when only 1 task was added
		select {
		case err := <-f.errc:
			return err
		default:
			return nil
		}
	}
}
