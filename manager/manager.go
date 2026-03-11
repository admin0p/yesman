package manager

import "sync"

type task func() int

type worker struct {
	id   int
	task task
	wg   *sync.WaitGroup
}

func (w *worker) ExeTask(finishCh chan<- int) {
	w.wg.Add(1)
	defer w.wg.Done()
	w.task()
	finishCh <- w.id
}

type pool interface {
	GetWorker() *worker
	AddWorker(w *worker)
}

type Manager struct {
	minWorkers int
	maxWorkers int
	taskCh     chan task
	finishCh   chan int
	workerPool pool
	wg         *sync.WaitGroup
}

func (m *Manager) Start() {
	//start min workers in idle pool
	for i := 0; i < m.minWorkers; i++ {
		m.workerPool.AddWorker(&worker{
			id: i,
			wg: m.wg,
		})
	}

	for t := range m.taskCh {
		// this should be a blocking call until a worker is available
		worker := m.workerPool.GetWorker()
		worker.task = t
		go worker.ExeTask(m.finishCh)

	}

}
