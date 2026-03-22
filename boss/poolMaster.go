package master

import (
	"context"
	"fmt"
	"sync"
	"yesman/worker"
)

// pool master is responsible to circulate workers within the pool
type PoolMaster struct {
	IdleWorker          []*worker.Worker
	ActiveWorker        []*worker.Worker
	finishCh            chan *worker.Worker
	idlePoolAvailableCh chan struct{}
	wg                  *sync.WaitGroup
}

func NewPoolMaster(ctx context.Context) *PoolMaster {

	pm := &PoolMaster{
		IdleWorker:          []*worker.Worker{},
		ActiveWorker:        []*worker.Worker{},
		finishCh:            make(chan *worker.Worker, 2),
		idlePoolAvailableCh: make(chan struct{}, 2),
		wg:                  &sync.WaitGroup{},
	}

	go pm.managerPool(pm.wg)
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

	return nil
}

// cleaner function the manages the transfer of active to idle pool when some worker finishes
func (pm *PoolMaster) managerPool(wg *sync.WaitGroup) {
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
		pm.idlePoolAvailableCh <- struct{}{}
	}
	//pm.idlePoolAvailableCh <- struct{}{}
	fmt.Println("DONE AND DUSTED")
}

func (pm *PoolMaster) Close() {
	close(pm.finishCh)
	// for finishEvent := range pm.finishCh {
	fmt.Println("DRAINING_FINISH_EVENT: worker ")
	// }
	pm.wg.Wait()
	fmt.Println("PM: waiting done")
	<-pm.idlePoolAvailableCh
	fmt.Println("DRAINING_IDLE_EVENT: ida ")
	close(pm.idlePoolAvailableCh)

}

func (pm *PoolMaster) GetAllWorker() []*worker.Worker {
	allWorkers := make([]*worker.Worker, 0, len(pm.ActiveWorker))
	allWorkers = append(allWorkers, pm.ActiveWorker...)
	return allWorkers
}

func (pm *PoolMaster) GetIdleWorkers() []*worker.Worker {
	allWorkers := make([]*worker.Worker, 0, len(pm.IdleWorker))
	allWorkers = append(allWorkers, pm.IdleWorker...)
	return allWorkers
}

func (pm *PoolMaster) GetIdlePushChan() <-chan struct{} {
	return pm.idlePoolAvailableCh
}
