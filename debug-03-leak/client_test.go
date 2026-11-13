package main

import (
	"context"
	"runtime"
	"testing"
	"time"
)

func TestFastestHeadNoLeak(t *testing.T) {
	c := &client{
		peers:   makePeers(time.Millisecond, 2*time.Millisecond, 3*time.Millisecond, 4*time.Millisecond),
		timeout: time.Second,
	}
	before := runtime.NumGoroutine()
	for range 10 {
		if _, err := c.fastestHead(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	time.Sleep(100 * time.Millisecond) // let the slower peers answer
	if after := runtime.NumGoroutine(); after > before {
		t.Fatalf("goroutines: %d before, %d after", before, after)
	}
}
