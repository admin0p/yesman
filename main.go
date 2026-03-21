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

	newWm := master.NewWorkerManager(ctx, 0, 1)
	newWm.Start()

	// cerate task
	for i := 0; i < 2; i++ {
		t := func() int {
			return i
		}
		newWm.PushTask(t)
	}
	fmt.Println("MAIN: closing")
	newWm.Stop()
}
