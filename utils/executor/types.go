package executor

//Task is the basic task that gets executed in executor
type Task func() error

//Executor is the interface for a basic executor pipeline
type Executor interface {
	//Add adds a task to executor queue
	Add(task Task)
	//Wait waits for all executors to finish or one of them to error based on option selected
	Wait() error
}

//Option represents different options available for Executor
type Option func(*config)
