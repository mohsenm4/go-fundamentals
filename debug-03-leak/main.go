// Command headwatch polls a set of peers for the chain head many times a
// second, keeps the best head it has seen, and hands new heads to a pool of
// verifiers.
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"time"
)

func verifier(heads <-chan head, verified *atomic.Uint64, wg *sync.WaitGroup) {
	defer wg.Done()
	for range heads {
		time.Sleep(time.Millisecond) // check the header chain
		verified.Add(1)
	}
}

func main() {
	every := flag.Duration("every", 50*time.Millisecond, "how often to poll the peers")
	workers := flag.Int("workers", 8, "verifier goroutines")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		t := time.NewTicker(time.Second)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				chainHeight.Add(1)
			case <-ctx.Done():
				return
			}
		}
	}()

	c := &client{
		peers: makePeers(
			10*time.Millisecond, 30*time.Millisecond, 60*time.Millisecond,
			100*time.Millisecond, 200*time.Millisecond, 400*time.Millisecond,
		),
		timeout: 120 * time.Millisecond,
	}

	heads := make(chan head, 16)
	var verified atomic.Uint64
	var wg sync.WaitGroup
	for range *workers {
		wg.Add(1)
		go verifier(heads, &verified, &wg)
	}

	poll := time.NewTicker(*every)
	defer poll.Stop()
	status := time.NewTicker(5 * time.Second)
	defer status.Stop()

	var best head
	var polls, failed int
	for {
		select {
		case <-poll.C:
			polls++
			h, err := c.fastestHead(ctx)
			if err != nil {
				failed++
				continue
			}
			if h.Height > best.Height {
				best = h
				select {
				case heads <- h:
				default: // verifiers are busy; the next head will do
				}
			}
		case <-status.C:
			log.Printf("polls=%d failed=%d best=%d (%s) verified=%d",
				polls, failed, best.Height, best.From, verified.Load())
		case <-ctx.Done():
			close(heads)
			wg.Wait()
			log.Printf("stopped after %d polls", polls)
			return
		}
	}
}
