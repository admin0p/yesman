package uppermanagement

import (
	"sync"
	"yesman/worker"
)

type WorkerPool interface {
	GetFinishCh() chan<- *worker.Worker
	GetWorker(maxW int) *worker.Worker
	AddWorker(w *worker.Worker)
	Close()
}

func NewWorkerManager(minW int, maxW int) *WorkerManager {
	return &WorkerManager{
		minWorker: minW,
		maxWorker: maxW,
		wg:        &sync.WaitGroup{},
		TaskChan:  make(chan worker.Task),
	}
}

type WorkerManager struct {
	maxWorker  int
	minWorker  int
	wg         *sync.WaitGroup
	WorkerPool WorkerPool
	TaskChan   chan worker.Task
}

func (wm *WorkerManager) Start() error {
	autoIncrId := 0
	wm.WorkerPool = &PoolMaster{
		finishCh:     make(chan *worker.Worker),
		IdleWorker:   make([]*worker.Worker, 0),
		ActiveWorker: make([]*worker.Worker, 0),
	}

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
		// close. the finish channels
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
