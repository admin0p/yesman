package main

import (
	"fmt"
	"sync"
)

type WorkerManager struct {
	taskQ     chan func() int
	workerNum int
	wg        *sync.WaitGroup
}

func (wm *WorkerManager) Start() {
	wm.taskQ = make(chan func() int, 4)
	wm.wg = &sync.WaitGroup{}
	for i := 0; i < wm.workerNum; i++ {
		wm.wg.Add(1)
		go wm.Run()
	}
}

func (wm *WorkerManager) AddTask(task func() int) {
	wm.taskQ <- task
}

func (wm *WorkerManager) Run() {
	defer wm.wg.Done()
	for task := range wm.taskQ {
		val := task()
		fmt.Println("task result:", val)
	}
}

func (wm *WorkerManager) Stop() {
	close(wm.taskQ)
	wm.wg.Wait()
}

func main() {
	wm := &WorkerManager{workerNum: 4}
	wm.Start()
	for i := 0; i < 10; i++ {
		task := func() int {
			return 2 * i
		}
		wm.AddTask(task)
	}
	wm.Stop()
}
