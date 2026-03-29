package worker

import (
	"fmt"

	"github.com/rs/xid"
)

type Task interface {
	Exec() error
	RetryExec() error
	GetIdentifier() int
}

type WorkerRules interface {
	GetRetryCount() int
	GetExpoBackOff() int
	GetErrorChan() chan<- error
}

type Worker struct {
	id      string
	task    Task
	myRules WorkerRules
}

func NewWorker(rules WorkerRules) *Worker {
	return &Worker{id: xid.New().String(), myRules: rules}
}

func (w *Worker) AssignTask(task Task) {
	w.task = task
}

func (w *Worker) Run(errCh chan<- error) (int, *Worker) {
	err := w.task.Exec()
	if err != nil {
		retryIndex := 0
		for err != nil && retryIndex < w.myRules.GetRetryCount() {
			fmt.Printf("WORKER %s: Task %d failed with error: %v. Retrying (%d/%d)...\n", w.id, w.task.GetIdentifier(), err, retryIndex+1, w.myRules.GetRetryCount())
			err = w.task.RetryExec()
			retryIndex++
		}
		if err != nil {
			fmt.Printf("WORKER %s: Task %d failed after %d retries. Error: %v\n", w.id, w.task.GetIdentifier(), retryIndex, err)
			w.myRules.GetErrorChan() <- fmt.Errorf("WORKER %s: Task %d failed after %d retries. Error: %v", w.id, w.task.GetIdentifier(), retryIndex, err)
		}
	}
	//time.Sleep( * time.Second)

	return w.task.GetIdentifier(), w
}

func (w *Worker) GetId() string {
	return w.id
}
