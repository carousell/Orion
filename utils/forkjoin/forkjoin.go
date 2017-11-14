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
	wg   sync.WaitGroup
	errc chan error
	done chan bool
}

func (f *forkJoin) Add(work Work) {
	f.wg.Add(1)
	go f.startWork(work)
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
		return err
	case <-f.done:
		return nil
	}
	return nil
}
