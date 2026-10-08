package main

import (
	"fmt"
	"testing"
)

func TestProcess(t *testing.T) {
	l := NewLedger()
	var deposits []Deposit
	for i := 0; i < 10000; i++ {
		deposits = append(deposits, Deposit{User: fmt.Sprintf("user-%d", i%10), Amount: 1})
	}
	process(l, deposits, 8)
	if got := l.Credited(); got != 10000 {
		t.Fatalf("credited = %d, want 10000", got)
	}
	if got := l.Balance("user-0"); got != 1000 {
		t.Fatalf("balance(user-0) = %d, want 1000", got)
	}
}
