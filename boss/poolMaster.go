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
type Busybody struct {
	IdleWorker          []*worker.Worker
	ActiveWorker        []*worker.Worker
	finishCh            chan *worker.Worker
	idlePoolAvailableCh chan struct{}
	wg                  *sync.WaitGroup
}

func NewBusybody(bufferSize int) *Busybody {

	bb := &Busybody{
		IdleWorker:          []*worker.Worker{},
		ActiveWorker:        []*worker.Worker{},
		finishCh:            make(chan *worker.Worker, bufferSize),
		idlePoolAvailableCh: make(chan struct{}),
		wg:                  &sync.WaitGroup{},
	}

	go bb.managePool(bb.wg)
	return bb
}

// returns the finish chan so that workers can signal their completion
// the worker push their task results here
func (bb *Busybody) GetFinishCh() chan<- *worker.Worker {
	return bb.finishCh
}

func (bb *Busybody) AddWorker(w *worker.Worker) {
	bb.IdleWorker = append(bb.IdleWorker, w)
}

func (bb *Busybody) getWorkerFromIdle() *worker.Worker {
	if len(bb.IdleWorker) == 0 {
		return nil
	}
	w := bb.IdleWorker[0]
	if w == nil {
		return nil
	}
	bb.IdleWorker = bb.IdleWorker[1:]
	bb.ActiveWorker = append(bb.ActiveWorker, w)
	return w
}

// this function retrieves an available worker from the pool or
// creates a new one if the maximum limit is not reached
// If the maximum limit is reached, it waits for a worker to finish and reuses it.
func (bb *Busybody) GetWorker(maxWorker int) *worker.Worker {
	if len(bb.IdleWorker) > 0 {
		return bb.getWorkerFromIdle()
	}

	if len(bb.IdleWorker)+len(bb.ActiveWorker) < maxWorker {
		w := worker.NewWorker()
		bb.ActiveWorker = append(bb.ActiveWorker, w)
		return w
	}

	for {
		time.Sleep(10 * time.Millisecond)
		if w := bb.getWorkerFromIdle(); w != nil {
			return w
		}
	}

}

// cleaner function the manages the transfer of active to idle pool when some worker finishes
func (bb *Busybody) managePool(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	for w := range bb.finishCh {
		for i, activeWorker := range bb.ActiveWorker {
			if activeWorker == w {
				bb.ActiveWorker = append(bb.ActiveWorker[:i], bb.ActiveWorker[i+1:]...)
				newWorker := worker.NewWorker()
				bb.IdleWorker = append(bb.IdleWorker, newWorker)
				break
			}
		}
	}
	close(bb.idlePoolAvailableCh)
	fmt.Println("BUSYBODY: DONE AND DUSTED")
}

// Drains the finish channel and waits for the managerPool function to finish
func (bb *Busybody) Close() {
	close(bb.finishCh)
	for finishEvent := range bb.finishCh {
		fmt.Println("DRAINING_FINISH_EVENT: worker ", finishEvent.GetId())
	}
	bb.wg.Wait()

}
