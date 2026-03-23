package yesman

import (
	"errors"
	"fmt"
	"sync"
	"yesman/worker"
)

// The YesMan
// Yes man manager (YesManManager) is your typical yes man who never says no to anything
// and does not do the work themselves and delegates to the poor workers :(
// this thing also does the same thing -
// this might look stupid in real life but unlike humans computers are purely emotionless so they don't care
// so this is quiet an interesting system in computer world
type YesManManager struct {
	maxWorker int
	minWorker int

	wg *sync.WaitGroup

	WorkerPool WorkerManager
	TaskChan   chan worker.Task
}

// this is the biggest tool or asset to the yes men
// like in our daily corporate life there is always a person who like a loyal dog to the yes man
// this is the lOYAL DOG!!
//
//	This is basically an interface or traits or personality expected from the "LOYAL DOG"
//
// I should have kept the name as LoyalDogTraits but that will confuse people so keeping this more understandable
// So this basically like that irritating guy in the office who always snitches about other people
// similarly this tracks the worker who is idle and when manager asks for a free worker it "snitches" on them
type WorkerManager interface {
	GetFinishCh() chan<- *worker.Worker
	GetWorker(maxW int) *worker.Worker
	AddWorker(w *worker.Worker)
	Close()
}

// Gives a new yes man
func NewYesMan(minW int, maxW int, poolMaster WorkerManager) *YesManManager {

	return &YesManManager{
		minWorker:  minW,
		maxWorker:  maxW,
		WorkerPool: poolMaster,
		wg:         &sync.WaitGroup{},
		TaskChan:   make(chan worker.Task),
	}
}

func (yesMan *YesManManager) Start() error {

	if yesMan.WorkerPool == nil {
		yesMan.WorkerPool = NewBusybody(yesMan.maxWorker)
	}

	for i := 0; i < yesMan.minWorker; i++ {

		w := worker.NewWorker()
		yesMan.WorkerPool.AddWorker(w)
	}

	yesMan.wg.Go(func() {
		for t := range yesMan.TaskChan {

			scapeGoat := yesMan.WorkerPool.GetWorker(yesMan.maxWorker)
			if scapeGoat == nil {
				panic(errors.New("Got a nil worker"))
			}

			scapeGoat.AssignTask(t)
			yesMan.wg.Add(1)

			go func(w *worker.Worker) {

				fmt.Println("YES_MAN: running worker ", w.GetId())
				res, goodGoat := w.Run()
				yesMan.WorkerPool.GetFinishCh() <- goodGoat
				fmt.Println("YES_MAN: worker ", goodGoat.GetId(), " finished task with result ", res)
				yesMan.wg.Done()

			}(scapeGoat)

		}
	})

	return nil
}

func (yesMan *YesManManager) PushTask(t worker.Task) {
	yesMan.TaskChan <- t
}

func (yesMan *YesManManager) Stop() error {
	close(yesMan.TaskChan)
	yesMan.wg.Wait()
	yesMan.WorkerPool.Close()
	return nil
}
