package master

import (
	"fmt"
	"os"
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
type YesManManager struct {
	maxWorker int
	minWorker int

	wg *sync.WaitGroup

	WorkerPool WorkerPool
	TaskChan   chan worker.Task
}

func NewYesMan(minW int, maxW int) *YesManManager {

	return &YesManManager{
		minWorker: minW,
		maxWorker: maxW,
		wg:        &sync.WaitGroup{},
		TaskChan:  make(chan worker.Task),
	}
}

func (yesMan *YesManManager) Start() error {

	yesMan.WorkerPool = NewPoolMaster()

	for i := 0; i < yesMan.minWorker; i++ {

		w := worker.NewWorker()
		yesMan.WorkerPool.AddWorker(w)
	}
	yesMan.wg.Add(1)
	go func() {
		defer yesMan.wg.Done()
		for t := range yesMan.TaskChan {
			fmt.Println("MANAGER: trying to get worker ")
			scapeGoat := yesMan.WorkerPool.GetWorker(yesMan.maxWorker)
			if scapeGoat == nil {
				fmt.Println("HOGAYA :)")
				os.Exit(1)
			}
			fmt.Println("MANAGER: got worker ", scapeGoat.GetId())
			scapeGoat.AssignTask(t)
			yesMan.wg.Add(1)
			fmt.Println("ADDED")
			go func(w *worker.Worker) {

				fmt.Println("MANAGER-WORKER: running worker ", w.GetId())
				res, goodGoat := w.Run()

				ch := yesMan.WorkerPool.GetFinishCh()
				fmt.Println("WORKER : chan", ch)
				ch <- goodGoat
				fmt.Println("MANAGER-WORKER: worker ", goodGoat.GetId(), " finished task with result ", res)
				yesMan.wg.Done()

			}(scapeGoat)

		}
	}()

	return nil
}

func (yesMan *YesManManager) PushTask(t worker.Task) {
	yesMan.TaskChan <- t
}

func (yesMan *YesManManager) Stop() error {
	fmt.Println("MANAGER: CLOSINGr")
	close(yesMan.TaskChan)
	yesMan.wg.Wait()
	fmt.Println("WAITING DONE")
	yesMan.WorkerPool.Close()
	return nil
}
