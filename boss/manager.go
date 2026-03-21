package master

import (
	"sync"
	"yesman/worker"
)

// the roles that master expects from the worker pool
type WorkerPool interface {
	GetFinishCh() chan<- *worker.Worker
	GetWorker(maxW int) *worker.Worker
	AddWorker(w *worker.Worker)
	Close()
}

// This is the main orchestrator , this is responsible to control the task assignment to worker
type WorkerManager struct {
	maxWorker  int
	minWorker  int
	wg         *sync.WaitGroup
	WorkerPool WorkerPool
	TaskChan   chan worker.Task
}

func NewWorkerManager(minW int, maxW int) *WorkerManager {
	return &WorkerManager{
		minWorker: minW,
		maxWorker: maxW,
		wg:        &sync.WaitGroup{},
		TaskChan:  make(chan worker.Task),
	}
}

func (wm *WorkerManager) Start() error {
	autoIncrId := 0
	wm.WorkerPool = NewPoolMaster(wm.wg)

	for i := 0; i < wm.minWorker; i++ {
		// new worker
		autoIncrId++
		w := worker.NewWorker(autoIncrId)
		wm.WorkerPool.AddWorker(w)
	}

	go func() {
		for t := range wm.TaskChan {
			// this is a blocking call the scheduler take care of
			// freeing worker,
			// generating new worker, reusing worker
			scapeGoat := wm.WorkerPool.GetWorker(wm.maxWorker)
			scapeGoat.AssignTask(t)
			wm.wg.Add(1)
			go func(w *worker.Worker) {
				defer wm.wg.Done()
				_, goodGoat := w.Run()
				wm.WorkerPool.GetFinishCh() <- goodGoat
			}(scapeGoat)

		}

		// close. poolMaster
		wm.WorkerPool.Close()
	}()

	return nil
}

func (wm *WorkerManager) PushTask(t worker.Task) {
	wm.TaskChan <- t
}

func (wm *WorkerManager) Stop() error {
	close(wm.TaskChan)
	wm.wg.Wait()
	return nil
}
