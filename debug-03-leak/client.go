package main

import (
	"context"
	"time"
)

type client struct {
	peers   []*peer
	timeout time.Duration
}

// fastestHead asks every peer at once and returns the first answer. The
// poller only needs a fresh head quickly; the verifiers check it later.
func (c *client) fastestHead(ctx context.Context) (head, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	answers := make(chan head, len(c.peers))
	for _, p := range c.peers {
		go func() {
			h, err := p.head()
			if err != nil {
				return
			}
			answers <- h
		}()
	}

	select {
	case h := <-answers:
		return h, nil
	case <-ctx.Done():
		return head{}, ctx.Err()
	}
}
