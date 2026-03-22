package main

import (
	"fmt"
	master "yesman/boss"
)

func main() {
	fmt.Println("Starting Worker Manager...")

	yesMan := master.NewYesMan(1, 3, nil)
	yesMan.Start()

	// cerate task
	for i := 0; i < 10; i++ {
		t := func() int {
			return i
		}
		yesMan.PushTask(t)
	}
	fmt.Println("MAIN: closing")
	yesMan.Stop()
}
