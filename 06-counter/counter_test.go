package main

import (
	"sync"
	"testing"
)

func TestCounter(t *testing.T) {
	counter := &Counter{}
	var wg sync.WaitGroup

	const goroutines = 100
	const incrementsPerGoroutine = 1000
	const expected = goroutines * incrementsPerGoroutine

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < incrementsPerGoroutine; j++ {
				counter.Inc()
			}
		}()
	}

	wg.Wait()
	if counter.Value() != expected {
		t.Errorf("Expected counter value to be %d, got %d", expected, counter.Value())
	}
}
