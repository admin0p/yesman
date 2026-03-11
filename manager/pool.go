package manager

import "time"

type workerPool struct {
	idle []*worker
	busy []*worker
}

func (wp *workerPool) AddWorker(worker *worker) {
	wp.idle = append(wp.idle, worker)
}

func (wp *workerPool) GetWorker(maxWorkers int) *worker {

	for len(wp.busy) >= maxWorkers {
		// wait for a worker to be available in worker pool
		time.Sleep(100 * time.Millisecond)
	}

	if len(wp.idle) > 0 {
		worker := wp.idle[0]
		wp.idle = wp.idle[1:]
		wp.busy = append(wp.busy, worker)
		return worker
	}

	if len(wp.busy) < maxWorkers {
		newWorker := &worker{
			id: len(wp.busy) + len(wp.idle) + 1,
		}
		wp.busy = append(wp.busy, newWorker)
		return newWorker
	}

	return nil
}
