package main

import (
	"fmt"
	"runtime"
	"time"
)

func main() {

	done := make(chan struct{})

	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	start := time.Now()

	for i := 0; i < 1000000; i++ {
		go func() {
			<-done
		}()
	}

	fmt.Println(runtime.NumGoroutine())
	elapsed := time.Since(start)
	fmt.Println("time ", elapsed)

	var after runtime.MemStats
	runtime.ReadMemStats(&after)

	heapDelta := after.HeapAlloc - before.HeapAlloc
	SysDelta := after.Sys - before.Sys
	StackInuse := after.StackInuse - before.StackInuse
	fmt.Println("heap delta ", heapDelta)
	fmt.Println("sys delta ", SysDelta)
	fmt.Println("stack inuse delta ", StackInuse)

	close(done)
}
