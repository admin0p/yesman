package main

import (
	"fmt"
	"sync"
	"time"
)

type task func() int

type Worker struct {
	id       int
	task     task
	finishCh chan<- int
}

func (w *Worker) ExeTask() {
	w.task()
	fmt.Println("Worker", w.id, "is executing task")
	w.finishCh <- w.id
}

type Manager struct {
	minWorkers int
	maxWorkers int
	taskCh     chan task
	finishCh   chan int
	wg         *sync.WaitGroup
	workerPool map[string]*[]Worker
}

func NewManager(minWorkers, maxWorkers int) *Manager {

	pool := map[string]*[]Worker{
		"idle":    &[]Worker{},
		"working": &[]Worker{},
	}
	return &Manager{
		minWorkers: minWorkers,
		maxWorkers: maxWorkers,
		taskCh:     make(chan task),
		finishCh:   make(chan int),
		workerPool: pool,
		wg:         &sync.WaitGroup{},
	}
}

func (m *Manager) cleaner() {
	for id := range m.finishCh {
		fmt.Println("Worker", id, "has finished task")
		// Move worker back to idle pool
		for i, worker := range *m.workerPool["working"] {
			if worker.id == id {
				*m.workerPool["idle"] = append(*m.workerPool["idle"], worker)
				*m.workerPool["working"] = append((*m.workerPool["working"])[:i], (*m.workerPool["working"])[i+1:]...)
				break
			}
		}
	}
	m.wg.Done()
}

func (m *Manager) Start() {
	for i := 0; i < m.minWorkers; i++ {
		worker := Worker{
			id:       i,
			finishCh: m.finishCh,
		}

		*m.workerPool["idle"] = append(*m.workerPool["idle"], worker)
	}

	m.wg.Add(2)
	go m.cleaner()
	go func() {
		for t := range m.taskCh {
			// Find an idle worker
			// assign task to worker and move worker to working pool
			if len(*m.workerPool["idle"]) > 0 {
				worker := (*m.workerPool["idle"])[0]
				*m.workerPool["idle"] = (*m.workerPool["idle"])[1:]
				worker.task = t
				*m.workerPool["working"] = append(*m.workerPool["working"], worker)
				m.wg.Add(1)
				go worker.ExeTask()
			} else {
				fmt.Println("No idle workers available, task is waiting")
			}
		}
	}()
}

func (m *Manager) AddTask(t task) {
	m.taskCh <- t
}

func (m *Manager) Stop() {
	m.wg.Wait()
	close(m.taskCh)
	close(m.finishCh)

}

func main() {

	manager := NewManager(1, 3)
	manager.Start()

	for i := 0; i < 5; i++ {
		task := func() int {
			time.Sleep(500 * time.Millisecond)
			return i
		}

		manager.AddTask(task)
		fmt.Println("added task", i)
	}

	manager.Stop()
}
