package main

import (
	"fmt"
	master "yesman/boss"
)

type task struct {
	id     int
	params string
}

func NewTask(id int, params string) *task {
	return &task{id: id, params: params}
}

func (t *task) Exec() error {
	// Implement task execution logic here
	fmt.Println("task params ==> ", t.params)
	return nil
}

func (t *task) RetryExec() error {
	// Implement task retry logic here
	return nil
}

func (t *task) GetIdentifier() int {
	return t.id
}

func main() {
	fmt.Println("Starting Worker Manager...")

	yesMan := master.NewYesMan(1, 3, nil)
	yesMan.Start()

	// cerate task
	for i := 0; i < 10; i++ {
		t := NewTask(i, fmt.Sprintf(" :) %d", i))
		yesMan.PushTask(t)
	}
	fmt.Println("MAIN: closing")
	yesMan.Stop()
}
