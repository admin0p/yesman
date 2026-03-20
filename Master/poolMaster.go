package master

import (
	"fmt"
	"yesman/worker"
)

// pool master is responsible to circulate workers within the pool
type PoolMaster struct {
	IdleWorker   []*worker.Worker
	ActiveWorker []*worker.Worker
	finishCh     chan *worker.Worker
}

// returns the finish chan so that workers can signal their completion
func (pm *PoolMaster) GetFinishCh() chan<- *worker.Worker {
	return pm.finishCh
}

func (pm *PoolMaster) AddWorker(w *worker.Worker) {
	pm.IdleWorker = append(pm.IdleWorker, w)
}

// this function retrieves an available worker from the pool or
// creates a new one if the maximum limit is not reached
// If the maximum limit is reached, it waits for a worker to finish and reuses it.
func (pm *PoolMaster) GetWorker(MaxWorker int) *worker.Worker {
	if len(pm.IdleWorker) > 0 {
		w := pm.IdleWorker[0]
		pm.IdleWorker = pm.IdleWorker[1:]
		pm.ActiveWorker = append(pm.ActiveWorker, w)
		return w
	}

	if len(pm.IdleWorker)+len(pm.ActiveWorker) < MaxWorker {
		w := worker.NewWorker(len(pm.IdleWorker) + len(pm.ActiveWorker) + 1)
		pm.ActiveWorker = append(pm.ActiveWorker, w)
		return w
	}

	finishedWorker := <-pm.finishCh
	*finishedWorker = worker.Worker{}
	return finishedWorker
}

func (pm *PoolMaster) Close() {
	for w := range pm.finishCh {
		fmt.Println("CLEANER_CLOSE: cleaned worker ", w.GetId())
	}
	close(pm.finishCh)
}
