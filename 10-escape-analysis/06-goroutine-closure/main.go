package main

import (
	"fmt"
	"sync"
)

//go:noinline
func run() {
	var wg sync.WaitGroup
	msg := "hello"

	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println(msg)
	}()

	wg.Wait()
}

func main() {
	run()
}
