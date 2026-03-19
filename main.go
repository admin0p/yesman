package main

import (
	uppermanagement "yesman/boss"
)

func main() {

	// inti workerManager

	newWm := uppermanagement.NewWorkerManager(1, 3)
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
