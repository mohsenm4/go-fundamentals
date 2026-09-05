package boundedchannel

import (
	"testing"
	"time"
)

func TestSend(t *testing.T) {
	ch := NewBoundedChannel(2)
	go func() {
		ch.Send(1)
		ch.Send(2)
		ch.Send(3) // This will block until a Receive is called
	}()

	time.Sleep(100 * time.Millisecond) // Give some time for the goroutine to run

	if len(ch.buf) != 2 {
		t.Errorf("Expected buffer length 2, got %d", len(ch.buf))
	}

	value := ch.Receive()
	if value != 1 {
		t.Errorf("Expected value 1, got %d", value)
	}

	value = ch.Receive()
	if value != 2 {
		t.Errorf("Expected value 2, got %d", value)
	}

	value = ch.Receive()
	if value != 3 {
		t.Errorf("Expected value 3, got %d", value)
	}
}

func TestReceive(t *testing.T) {
	ch := NewBoundedChannel(2)
	go func() {
		time.Sleep(100 * time.Millisecond) // Ensure Receive is called first
		ch.Send(1)
		ch.Send(2)
	}()

	value := ch.Receive()
	if value != 1 {
		t.Errorf("Expected value 1, got %d", value)
	}

	value = ch.Receive()
	if value != 2 {
		t.Errorf("Expected value 2, got %d", value)
	}
}

func TestBoundedChannel(t *testing.T) {
	ch := NewBoundedChannel(2)

	// Test sending and receiving in a single goroutine
	ch.Send(1)
	ch.Send(2)

	if len(ch.buf) != 2 {
		t.Errorf("Expected buffer length 2, got %d", len(ch.buf))
	}

	value := ch.Receive()
	if value != 1 {
		t.Errorf("Expected value 1, got %d", value)
	}

	value = ch.Receive()
	if value != 2 {
		t.Errorf("Expected value 2, got %d", value)
	}
}
