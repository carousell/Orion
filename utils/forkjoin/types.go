package forkjoin

type Work func() error

type ForkJoin interface {
	Add(Work)
	Wait() error
}
