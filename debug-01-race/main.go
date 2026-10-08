package main

import (
	"fmt"
	"sync"
	"time"
)

type Deposit struct {
	User   string
	Amount int64
}

type Ledger struct {
	mu       sync.Mutex
	balances map[string]int64
	credited int // how many deposits were credited
}

func NewLedger() *Ledger {
	return &Ledger{balances: make(map[string]int64)}
}

func (l *Ledger) Credit(d Deposit) {
	l.mu.Lock()
	l.balances[d.User] += d.Amount
	l.credited++
	l.mu.Unlock()
}

func (l *Ledger) Credited() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.credited
}

func (l *Ledger) Balance(user string) int64 {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.balances[user]
}

// process credits all deposits with n workers and prints progress while it runs.
func process(l *Ledger, deposits []Deposit, n int) {
	ch := make(chan Deposit)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for d := range ch {
				l.Credit(d)
			}
		}()
	}

	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				fmt.Println("progress:", l.Credited())
			}
		}
	}()

	for _, d := range deposits {
		ch <- d
	}
	close(ch)
	wg.Wait()
	close(done)
}

func main() {
	l := NewLedger()
	var deposits []Deposit
	for i := 0; i < 10000; i++ {
		deposits = append(deposits, Deposit{User: fmt.Sprintf("user-%d", i%10), Amount: 1})
	}
	process(l, deposits, 8)
	fmt.Println("credited:", l.Credited(), "user-0:", l.Balance("user-0"))
}
