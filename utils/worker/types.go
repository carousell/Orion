package worker

import (
	"context"

	"github.com/RichardKnop/machinery/v1/tasks"
)

//Worker is the interface for worker
type Worker interface {
	Schedule(ctx context.Context, name, payload string, options ...ScheduleOption) error
	RegisterTask(name string, taskFunc Work) error
	RunWorker(name string, concurrency int)
	CloseWorker()
}

//Config is the config used to intialize workers
type Config struct {
	LocalMode    bool
	RabbitConfig *RabbitMQConfig
}

//RabbitMQConfig is the config used for scheduling tasks through rabbitmq
type RabbitMQConfig struct {
	UserName    string
	Password    string
	BrokerVHost string
	Host        string
	Port        string
	QueueName   string
}

//Work is the type of task that can be exeucted by Worker
type Work func(ctx context.Context, payload string) error
type wrappedWork func(payload string) error

//ScheduleConfig is the config used when scheduling a task
type ScheduleConfig struct {
	retries   int
	queueName string
}

//ScheduleOption represents different options available for Schedule
type ScheduleOption func(*ScheduleConfig)

type fakeBackend struct {
}

func (f *fakeBackend) InitGroup(groupUUID string, taskUUIDs []string) error {
	return nil
}

func (f *fakeBackend) GroupCompleted(groupUUID string, groupTaskCount int) (bool, error) {
	return true, nil
}

func (f *fakeBackend) GroupTaskStates(groupUUID string, groupTaskCount int) ([]*tasks.TaskState, error) {
	return []*tasks.TaskState{}, nil
}

func (f *fakeBackend) TriggerChord(groupUUID string) (bool, error) {
	return true, nil
}

func (f *fakeBackend) SetStatePending(signature *tasks.Signature) error {
	return nil
}

func (f *fakeBackend) SetStateReceived(signature *tasks.Signature) error {
	return nil
}

func (f *fakeBackend) SetStateStarted(signature *tasks.Signature) error {
	return nil
}

func (f *fakeBackend) SetStateRetry(signature *tasks.Signature) error {
	return nil
}

func (f *fakeBackend) SetStateSuccess(signature *tasks.Signature, results []*tasks.TaskResult) error {
	return nil
}

func (f *fakeBackend) SetStateFailure(signature *tasks.Signature, err string) error {
	return nil
}

func (f *fakeBackend) GetState(taskUUID string) (*tasks.TaskState, error) {
	return new(tasks.TaskState), nil
}

func (f *fakeBackend) PurgeState(taskUUID string) error {
	return nil
}

func (f *fakeBackend) PurgeGroupMeta(groupUUID string) error {
	return nil
}
