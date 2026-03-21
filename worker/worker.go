package worker

import (
	"fmt"
	"time"

	"github.com/rs/xid"
)

type Task func() int

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
	result := w.task()

	time.Sleep(2 * time.Second)

	return result, w
}

func (w *Worker) GetId() string {
	return w.id
}
