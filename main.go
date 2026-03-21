package main

import (
	"fmt"
	master "yesman/boss"
)

func main() {
	fmt.Println("Starting Worker Manager...")
	// inti workerManager

	newWm := master.NewWorkerManager(1, 3)
	newWm.Start()

	// cerate task
	for i := 0; i < 3; i++ {
		t := func() int {
			return i
		}

		newWm.PushTask(t)
	}

	newWm.Stop()
}
