package forkjoin

type Task func() error

type ForkJoin interface {
	Add(Task)
	Wait() error
}
