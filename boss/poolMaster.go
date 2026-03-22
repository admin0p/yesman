package master

import (
	"fmt"
	"sync"
	"time"
	"yesman/worker"
)

// The loyal dog to the master (Yes man)
// Like that every irritating person in office
// who pokes his nose into everyone's business but does not mind his own
// Similarly, this dude tracks all worker's activity and gives up the idle ones to master when he asks
type PoolMaster struct {
	IdleWorker          []*worker.Worker
	ActiveWorker        []*worker.Worker
	finishCh            chan *worker.Worker
	idlePoolAvailableCh chan struct{}
	wg                  *sync.WaitGroup
}

func NewPoolMaster() *PoolMaster {

	pm := &PoolMaster{
		IdleWorker:          []*worker.Worker{},
		ActiveWorker:        []*worker.Worker{},
		finishCh:            make(chan *worker.Worker, 2),
		idlePoolAvailableCh: make(chan struct{}),
		wg:                  &sync.WaitGroup{},
	}

	go pm.managePool(pm.wg)
	return pm
}

// returns the finish chan so that workers can signal their completion
// the worker push their task results here
func (pm *PoolMaster) GetFinishCh() chan<- *worker.Worker {
	return pm.finishCh
}

func (pm *PoolMaster) AddWorker(w *worker.Worker) {
	pm.IdleWorker = append(pm.IdleWorker, w)
}

func (pm *PoolMaster) getWorkerFromIdle() *worker.Worker {
	if len(pm.IdleWorker) == 0 {
		return nil
	}
	w := pm.IdleWorker[0]
	if w == nil {
		return nil
	}
	pm.IdleWorker = pm.IdleWorker[1:]
	pm.ActiveWorker = append(pm.ActiveWorker, w)
	return w
}

// this function retrieves an available worker from the pool or
// creates a new one if the maximum limit is not reached
// If the maximum limit is reached, it waits for a worker to finish and reuses it.
func (pm *PoolMaster) GetWorker(maxWorker int) *worker.Worker {
	if len(pm.IdleWorker) > 0 {
		//fmt.Println("SCHEDULER-1: from idle pool")
		return pm.getWorkerFromIdle()
	}

	if len(pm.IdleWorker)+len(pm.ActiveWorker) < maxWorker {
		w := worker.NewWorker()
		//fmt.Println("SCHEDULER-1: creating new worker ", w.GetId())
		pm.ActiveWorker = append(pm.ActiveWorker, w)
		return w
	}
	fmt.Println("Waiting idle worker ...")
	for {
		time.Sleep(10 * time.Millisecond)
		if w := pm.getWorkerFromIdle(); w != nil {
			return w
		}
	}

}

// cleaner function the manages the transfer of active to idle pool when some worker finishes
func (pm *PoolMaster) managePool(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	for w := range pm.finishCh {
		for i, activeWorker := range pm.ActiveWorker {
			if activeWorker == w {
				fmt.Println("Freeing worker ", activeWorker.GetId())
				pm.ActiveWorker = append(pm.ActiveWorker[:i], pm.ActiveWorker[i+1:]...)
				newWorker := worker.NewWorker()
				pm.IdleWorker = append(pm.IdleWorker, newWorker)
				break
			}
		}
	}
	close(pm.idlePoolAvailableCh)
	fmt.Println("DONE AND DUSTED")
}

func (pm *PoolMaster) Close() {
	close(pm.finishCh)
	for finishEvent := range pm.finishCh {
		fmt.Println("DRAINING_FINISH_EVENT: worker ", finishEvent.GetId())
	}
	pm.wg.Wait()
	fmt.Println("PM_CLOSE: waiting done")

}
