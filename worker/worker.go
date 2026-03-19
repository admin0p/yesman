package worker

import "fmt"

type Task func() int

type Worker struct {
	id   int
	task Task
}

func NewWorker(id int) *Worker {
	return &Worker{id: id}
}

func workerLog(message string) {
	fmt.Println("WORKER:{ ", message, " }")
}

func (w *Worker) AssignTask(task Task) {
	w.task = task
}

func (w *Worker) Run() (int, *Worker) {
	workerLog(fmt.Sprintf("worker %d received task", w.id))
	result := w.task()
	workerLog(fmt.Sprintf("Finished task with result := %d", result))
	return result, w
}

func (w *Worker) GetId() int {
	return w.id
}
