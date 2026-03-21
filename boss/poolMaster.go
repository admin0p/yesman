package master

import (
	"context"
	"fmt"
	"yesman/worker"
)

// pool master is responsible to circulate workers within the pool
type PoolMaster struct {
	IdleWorker          []*worker.Worker
	ActiveWorker        []*worker.Worker
	finishCh            chan *worker.Worker
	idlePoolAvailableCh chan struct{}
}

func NewPoolMaster(ctx context.Context) *PoolMaster {

	pm := &PoolMaster{
		IdleWorker:          []*worker.Worker{},
		ActiveWorker:        []*worker.Worker{},
		finishCh:            make(chan *worker.Worker),
		idlePoolAvailableCh: make(chan struct{}),
	}

	go pm.managerPool()
	return pm
}

// returns the finish chan so that workers can signal their completion
func (pm *PoolMaster) GetFinishCh() chan<- *worker.Worker {
	return pm.finishCh
}

func (pm *PoolMaster) AddWorker(w *worker.Worker) {
	pm.IdleWorker = append(pm.IdleWorker, w)
}

func (pm *PoolMaster) getWorkerFromIdle() *worker.Worker {
	w := pm.IdleWorker[0]
	pm.IdleWorker = pm.IdleWorker[1:]
	pm.ActiveWorker = append(pm.ActiveWorker, w)
	return w
}

// this function retrieves an available worker from the pool or
// creates a new one if the maximum limit is not reached
// If the maximum limit is reached, it waits for a worker to finish and reuses it.
func (pm *PoolMaster) GetWorker(ctx context.Context, maxWorker int) *worker.Worker {
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

	select {
	case w, ok := <-pm.finishCh:
		if !ok {
			return nil
		}
		fmt.Println("SCHEDULER: from finish task ", w.GetId())
		return w
	case _, ok := <-pm.idlePoolAvailableCh:
		if !ok {
			return nil
		}
		fmt.Println("SCHEDULER: from idle pool")
		return pm.getWorkerFromIdle()
	}

}

// cleaner function the manages the transfer of active to idle pool when some worker finishes
func (pm *PoolMaster) managerPool() {
	for w := range pm.finishCh {
		for i, activeWorker := range pm.ActiveWorker {
			if activeWorker == w {
				pm.ActiveWorker = append(pm.ActiveWorker[:i], pm.ActiveWorker[i+1:]...)
				break
			}
		}
		pm.idlePoolAvailableCh <- struct{}{}
	}
	fmt.Println("DONE AND DUSTED")
}

func (pm *PoolMaster) Close() {
	close(pm.finishCh)
	for finishEvent := range pm.finishCh {
		fmt.Println("DRAINING_FINISH_EVENT: worker ", finishEvent.GetId())
	}
	close(pm.idlePoolAvailableCh)
	for range pm.idlePoolAvailableCh {
		fmt.Println("DRAINING_IDLE")
	}
}

func (pm *PoolMaster) GetAllWorker() []*worker.Worker {
	allWorkers := make([]*worker.Worker, 0, len(pm.ActiveWorker))
	allWorkers = append(allWorkers, pm.ActiveWorker...)
	return allWorkers
}
