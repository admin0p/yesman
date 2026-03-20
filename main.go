package main

import master "yesman/Master"

func main() {

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
