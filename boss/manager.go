package master

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

	// retry count for a task in case of failure
	workerRules WorkerRules

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

// just a worker rules as to what a worker can do
// This should match the capabilities of the worker
// checkout the poor worker code as to what they can do and what they can't do
// NOTE: even though a yesman manager is an emotionless machine,
// it should not assign a task to a poor worker that is not capable of doing it :)
type WorkerRules struct {
	RetryCount  int
	ExpoBackOff int
	// channel to send task that failed to execute after retrying
	ErrorChan chan error
}

func (wr WorkerRules) GetRetryCount() int {
	return wr.RetryCount
}
func (wr WorkerRules) GetExpoBackOff() int {
	return wr.ExpoBackOff
}
func (wr WorkerRules) GetErrorChan() chan<- error {
	return wr.ErrorChan
}

// Gives a new yes man
func NewYesMan(minW int, maxW int, poolMaster WorkerManager, errCh chan error) *YesManManager {

	return &YesManManager{
		minWorker:   minW,
		maxWorker:   maxW,
		WorkerPool:  poolMaster,
		workerRules: WorkerRules{RetryCount: 3, ExpoBackOff: 2, ErrorChan: errCh},
		wg:          &sync.WaitGroup{},
		TaskChan:    make(chan worker.Task),
	}
}

func (yesMan *YesManManager) Start() error {

	if yesMan.WorkerPool == nil {
		yesMan.WorkerPool = NewBusybody(yesMan.maxWorker, yesMan.workerRules)
	}

	for i := 0; i < yesMan.minWorker; i++ {

		w := worker.NewWorker(yesMan.workerRules)
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
				res, goodGoat := w.Run(yesMan.workerRules.GetErrorChan())
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
