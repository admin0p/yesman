package master

import (
	"context"
	"fmt"
	"os"
	"sync"
	"yesman/worker"
)

// the roles that master expects from the worker pool
type WorkerPool interface {
	GetFinishCh() chan<- *worker.Worker
	GetIdlePushChan() <-chan struct{}
	GetWorker(maxW int) *worker.Worker
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
	wm.wg.Add(1)
	go func() {
		defer wm.wg.Done()
		for t := range wm.TaskChan {
			fmt.Println("MANAGER: trying to get worker ")
			scapeGoat := wm.WorkerPool.GetWorker(wm.maxWorker)
			if scapeGoat == nil {
				fmt.Println("HOGAYA :)")
				os.Exit(1)
			}
			fmt.Println("MANAGER: got worker ", scapeGoat.GetId())
			scapeGoat.AssignTask(t)
			wm.wg.Add(1)
			fmt.Println("ADDED")
			go func(w *worker.Worker) {

				fmt.Println("MANAGER-WORKER: running worker ", w.GetId())
				res, goodGoat := w.Run()

				ch := wm.WorkerPool.GetFinishCh()
				fmt.Println("WORKER : chan", ch)
				ch <- goodGoat
				fmt.Println("MANAGER-WORKER: worker ", goodGoat.GetId(), " finished task with result ", res)
				wm.wg.Done()

			}(scapeGoat)

		}
	}()

	return nil
}

func (wm *WorkerManager) PushTask(t worker.Task) {
	wm.TaskChan <- t
}

func (wm *WorkerManager) Stop() error {
	fmt.Println("MANAGER: CLOSINGr")
	close(wm.TaskChan)
	wm.wg.Wait()
	fmt.Println("WAITING DONE")
	wm.WorkerPool.Close()
	return nil
}
