package main

import (
	"context"
	"fmt"
	master "yesman/boss"
)

func main() {
	fmt.Println("Starting Worker Manager...")
	// inti workerManager
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	newWm := master.NewWorkerManager(ctx, 1, 3)
	newWm.Start()

	// cerate task
	for i := 0; i < 10; i++ {
		t := func() int {
			return i
		}
		newWm.PushTask(t)
	}
	fmt.Println("MAIN: closing")
	newWm.Stop()
}
