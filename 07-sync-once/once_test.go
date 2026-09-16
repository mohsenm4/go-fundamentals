package once

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestBrokenOnce_ReturnsBeforeInitDone(t *testing.T) {
	var o BrokenOnce
	var ready atomic.Bool
	var sawNotReady atomic.Int32

	f := func() {
		time.Sleep(50 * time.Millisecond)
		ready.Store(true)
	}

	var wg sync.WaitGroup
	wg.Add(100)

	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()

			o.Do(f)

			if !ready.Load() {
				sawNotReady.Add(1)
			}
		}()
	}

	wg.Wait()

	t.Logf("returned before init finished: %d / 100", sawNotReady.Load())
}
func TestOnce_WaitsForInit(t *testing.T) {
	var o Once
	var ready atomic.Bool
	var sawNotReady atomic.Int32
	f := func() {
		time.Sleep(50 * time.Millisecond)
		ready.Store(true)
	}
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			o.Do(f)
			if !ready.Load() {
				sawNotReady.Add(1)
			}
		}()
	}
	wg.Wait()
	if n := sawNotReady.Load(); n != 0 {
		t.Fatalf("%d goroutines returned before init finished", n)
	}
}

func TestBothRunExactlyOnce(t *testing.T) {
	run := func(t *testing.T, do func(func())) {
		var calls atomic.Int32
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				do(func() { calls.Add(1) })
			}()
		}
		wg.Wait()
		if calls.Load() != 1 {
			t.Fatalf("f ran %d times, want 1", calls.Load())
		}
	}
	t.Run("broken", func(t *testing.T) { var o BrokenOnce; run(t, o.Do) })
	t.Run("correct", func(t *testing.T) { var o Once; run(t, o.Do) })
}
