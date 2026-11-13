package main

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"sync/atomic"
	"time"
)

type head struct {
	Height uint64
	From   string
}

var errUnavailable = errors.New("peer unavailable")

// chainHeight is the height of the simulated chain. It grows by one a second.
var chainHeight atomic.Uint64

// peer stands in for a remote node. head is a blocking call over the network,
// and the client library we use for it takes no context.
type peer struct {
	name    string
	latency time.Duration // average round trip
}

func (p *peer) head() (head, error) {
	time.Sleep(p.latency/2 + rand.N(p.latency))
	if rand.IntN(20) == 0 {
		return head{}, errUnavailable
	}
	lag := uint64(rand.IntN(3))
	h := chainHeight.Load()
	if h > lag {
		h -= lag
	}
	return head{Height: h, From: p.name}, nil
}

func makePeers(latencies ...time.Duration) []*peer {
	ps := make([]*peer, len(latencies))
	for i, l := range latencies {
		ps[i] = &peer{name: fmt.Sprintf("peer-%d", i), latency: l}
	}
	return ps
}
