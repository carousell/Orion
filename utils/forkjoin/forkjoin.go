package forkjoin

import (
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
	mu      sync.Mutex
	errored bool
}

func (f *forkJoin) Add(work Work) {
	f.mu.Lock()
	defer f.mu.Unlock()
	// if we have already errored dont start work
	if !f.errored {
		f.wg.Add(1)
		go f.startWork(work)
	}
}

func (f *forkJoin) startWork(work Work) {
	defer f.wg.Done()
	err := work()
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
			f.mu.Lock()
			f.errored = true
			f.mu.Unlock()
		}
		return err
	case <-f.done:
		return nil
	}
	return nil
}
