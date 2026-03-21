package master

import (
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
}

func NewPoolMaster(wg *sync.WaitGroup) *PoolMaster {

	pm := &PoolMaster{
		IdleWorker:          []*worker.Worker{},
		ActiveWorker:        []*worker.Worker{},
		finishCh:            make(chan *worker.Worker),
		idlePoolAvailableCh: make(chan struct{}),
	}
	wg.Add(1)
	defer wg.Done()

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
func (pm *PoolMaster) GetWorker(MaxWorker int) *worker.Worker {
	if len(pm.IdleWorker) > 0 {
		return pm.getWorkerFromIdle()
	}

	if len(pm.IdleWorker)+len(pm.ActiveWorker) < MaxWorker {
		w := worker.NewWorker(len(pm.IdleWorker) + len(pm.ActiveWorker) + 1)
		pm.ActiveWorker = append(pm.ActiveWorker, w)
		return w
	}

	select {
	case w := <-pm.finishCh:
		return w
	case <-pm.idlePoolAvailableCh:
		return pm.getWorkerFromIdle()
	}

}

// cleaner function the manages the transfer of active to idle pool when some worker finishes
func (pm *PoolMaster) managerPool() {

	for w := range pm.finishCh {
		for i, activeWorker := range pm.ActiveWorker {
			fmt.Println("CLEANER: moving worker ", w.GetId(), " from active to idle")
			if activeWorker == w {
				pm.ActiveWorker = append(pm.ActiveWorker[:i], pm.ActiveWorker[i+1:]...)
				break
			}
		}
		pm.IdleWorker = append(pm.IdleWorker, w)
		pm.idlePoolAvailableCh <- struct{}{}
	}

}

func (pm *PoolMaster) Close() {
	for w := range pm.finishCh {
		fmt.Println("CLEANER_CLOSE: cleaned worker ", w.GetId())
	}
	close(pm.finishCh)
}
