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

type Worker struct {
	id   string
	task Task
}

func NewWorker() *Worker {
	return &Worker{id: xid.New().String()}
}

func workerLog(message string) {
	fmt.Println("WORKER:{ ", message, " }")
}

func (w *Worker) AssignTask(task Task) {
	w.task = task
}

func (w *Worker) Run() (int, *Worker) {
	err := w.task.Exec()
	if err != nil {
		fmt.Println("error in task")
	}
	//time.Sleep( * time.Second)

	return w.task.GetIdentifier(), w
}

func (w *Worker) GetId() string {
	return w.id
}
