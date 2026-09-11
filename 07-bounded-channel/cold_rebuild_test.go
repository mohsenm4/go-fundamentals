package boundedchannel

import (
	"sync"
	"testing"
	"time"
)

func TestSendAndRecv(t *testing.T) {
	ch := New[int](3)

	ch.Send(1)
	ch.Send(2)
	ch.Send(3)

	tests := []int{1, 2, 3}

	for _, want := range tests {
		got, ok := ch.Recv()

		if !ok {
			t.Fatalf("expected ok=true")
		}

		if got != want {
			t.Fatalf("expected %d, got %d", want, got)
		}
	}
}

func TestRecvBlocksWhenEmpty(t *testing.T) {
	ch := New[int](1)

	done := make(chan struct{})

	go func() {
		v, ok := ch.Recv()

		if !ok {
			t.Errorf("expected ok=true")
		}

		if v != 42 {
			t.Errorf("expected 42, got %d", v)
		}

		close(done)
	}()

	select {
	case <-done:
		t.Fatal("Recv should block while buffer is empty")

	case <-time.After(50 * time.Millisecond):
	}

	ch.Send(42)

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Recv did not unblock after Send")
	}
}

func TestSendBlocksWhenFull(t *testing.T) {
	ch := New[int](1)

	ch.Send(1)

	done := make(chan struct{})

	go func() {
		ch.Send(2)
		close(done)
	}()

	select {
	case <-done:
		t.Fatal("Send should block when buffer is full")

	case <-time.After(50 * time.Millisecond):
	}

	v, ok := ch.Recv()

	if !ok {
		t.Fatal("expected ok=true")
	}

	if v != 1 {
		t.Fatalf("expected 1, got %d", v)
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Send did not unblock after Recv")
	}

	v, ok = ch.Recv()

	if !ok {
		t.Fatal("expected ok=true")
	}

	if v != 2 {
		t.Fatalf("expected 2, got %d", v)
	}
}

func TestCircularBuffer(t *testing.T) {
	ch := New[int](3)

	ch.Send(1)
	ch.Send(2)
	ch.Send(3)

	v, _ := ch.Recv()
	if v != 1 {
		t.Fatalf("expected 1, got %d", v)
	}

	v, _ = ch.Recv()
	if v != 2 {
		t.Fatalf("expected 2, got %d", v)
	}

	ch.Send(4)
	ch.Send(5)

	v, _ = ch.Recv()
	if v != 3 {
		t.Fatalf("expected 3, got %d", v)
	}

	v, _ = ch.Recv()
	if v != 4 {
		t.Fatalf("expected 4, got %d", v)
	}

	v, _ = ch.Recv()
	if v != 5 {
		t.Fatalf("expected 5, got %d", v)
	}
}

func TestClose(t *testing.T) {
	ch := New[int](3)

	ch.Send(1)
	ch.Send(2)

	ch.Close()

	v, ok := ch.Recv()
	if !ok || v != 1 {
		t.Fatalf("expected 1, true; got %d, %v", v, ok)
	}

	v, ok = ch.Recv()
	if !ok || v != 2 {
		t.Fatalf("expected 2, true; got %d, %v", v, ok)
	}

	v, ok = ch.Recv()

	if ok {
		t.Fatal("expected ok=false after closed channel is drained")
	}

	if v != 0 {
		t.Fatalf("expected zero value, got %d", v)
	}
}

func TestSendAfterClosePanics(t *testing.T) {
	ch := New[int](1)

	ch.Close()

	defer func() {
		if recover() == nil {
			t.Fatal("expected Send after Close to panic")
		}
	}()

	ch.Send(1)
}

func TestCloseUnblocksWaitingReceiver(t *testing.T) {
	ch := New[int](1)

	done := make(chan struct{})

	go func() {
		_, ok := ch.Recv()

		if ok {
			t.Error("expected ok=false")
		}

		close(done)
	}()

	select {
	case <-done:
		t.Fatal("Recv should still be blocked")

	case <-time.After(50 * time.Millisecond):
	}

	ch.Close()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Recv did not unblock after Close")
	}
}

func TestCloseUnblocksWaitingSender(t *testing.T) {
	ch := New[int](1)

	ch.Send(1)

	done := make(chan struct{})

	go func() {
		defer func() {
			if recover() == nil {
				t.Error("expected Send to panic after Close")
			}

			close(done)
		}()

		ch.Send(2)
	}()

	select {
	case <-done:
		t.Fatal("Send should still be blocked")

	case <-time.After(50 * time.Millisecond):
	}

	ch.Close()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Send did not unblock after Close")
	}
}

func TestCloseIsIdempotent(t *testing.T) {
	ch := New[int](1)

	ch.Close()
	ch.Close()
	ch.Close()

	_, ok := ch.Recv()

	if ok {
		t.Fatal("expected closed channel")
	}
}

func TestConcurrentSendRecv(t *testing.T) {
	ch := New[int](10)

	const count = 1000

	var wg sync.WaitGroup

	received := make(chan int, count)

	for i := 0; i < 5; i++ {
		wg.Add(1)

		go func(start int) {
			defer wg.Done()

			for j := 0; j < count/5; j++ {
				ch.Send(start + j)
			}
		}(i * 1000)
	}

	for i := 0; i < 5; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for j := 0; j < count/5; j++ {
				v, ok := ch.Recv()

				if !ok {
					t.Errorf("unexpected ok=false")
					return
				}

				received <- v
			}
		}()
	}

	wg.Wait()

	close(received)

	if len(received) != count {
		t.Fatalf("expected %d received values, got %d", count, len(received))
	}
}
