package master

import (
	"context"
	"fmt"
	"sync"
	"yesman/worker"
)

// the roles that master expects from the worker pool
type WorkerPool interface {
	GetFinishCh() chan<- *worker.Worker
	GetWorker(ctx context.Context, maxW int) *worker.Worker
	AddWorker(w *worker.Worker)
	GetAllWorker() []*worker.Worker //test function
	Close()
}

// This is the main orchestrator , this is responsible to control the task assignment to worker
type WorkerManager struct {
	maxWorker int
	minWorker int

	wg  *sync.WaitGroup
	ctx context.Context
	can context.CancelFunc

	WorkerPool WorkerPool
	TaskChan   chan worker.Task
}

func NewWorkerManager(ctx context.Context, minW int, maxW int) *WorkerManager {
	cancelableCtx, can := context.WithCancel(ctx)
	return &WorkerManager{
		minWorker: minW,
		maxWorker: maxW,
		wg:        &sync.WaitGroup{},
		TaskChan:  make(chan worker.Task),
		ctx:       cancelableCtx,
		can:       can,
	}
}

func (wm *WorkerManager) Start() error {

	wm.WorkerPool = NewPoolMaster(wm.ctx)

	for i := 0; i < wm.minWorker; i++ {

		w := worker.NewWorker()
		wm.WorkerPool.AddWorker(w)
	}

	go func() {
		for t := range wm.TaskChan {
			fmt.Println("MANAGER: trying to get worker ")
			scapeGoat := wm.WorkerPool.GetWorker(wm.ctx, wm.maxWorker)
			if scapeGoat == nil {
				fmt.Println("MANAGER: received nil worker")
				continue
			}
			fmt.Println("MANAGER: got worker ", scapeGoat.GetId())
			scapeGoat.AssignTask(t)
			wm.wg.Add(1)
			go func(w *worker.Worker) {

				fmt.Println("MANAGER-WORKER: running worker ", w.GetId())
				res, goodGoat := w.Run()

				wm.WorkerPool.GetFinishCh() <- goodGoat

				fmt.Println("MANAGER-WORKER: worker ", goodGoat.GetId(), " finished task with result ", res)
				wm.wg.Done()

			}(scapeGoat)

		}

		// close. poolMaster
		// wm.WorkerPool.Close()
	}()

	return nil
}

func (wm *WorkerManager) PushTask(t worker.Task) {
	wm.TaskChan <- t
}

func (wm *WorkerManager) Stop() error {
	close(wm.TaskChan)
	wm.wg.Wait()
	wm.WorkerPool.Close()
	return nil
}
