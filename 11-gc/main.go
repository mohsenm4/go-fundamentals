package main

import (
	"fmt"
	"runtime"
	"time"
)

// live stays reachable for the whole run: ~10 MB live heap.
var live [][]byte

// sink makes each 1 KB buffer escape; the previous one becomes garbage.
var sink []byte

func main() {
	for i := 0; i < 10_000; i++ {
		live = append(live, make([]byte, 1024))
	}

	start := time.Now()
	for i := 0; i < 1_000_000; i++ {
		sink = make([]byte, 1024)
	}
	elapsed := time.Since(start)

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("NumGC=%d time=%v\n", m.NumGC, elapsed)
}
