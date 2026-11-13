// Command blockidx follows a simulated chain. It fetches every block, keeps
// an index of the headers for lookups, and keeps the newest raw blocks in a
// small cache for peers.
package main

import (
	"bytes"
	"flag"
	"log"
	"os"
	"os/signal"
	"time"
)

func main() {
	tip := flag.Uint64("tip", 10_000, "height of the chain tip when we start")
	cacheSize := flag.Int("cache", 64, "raw blocks kept in memory")
	flag.Parse()

	ix := newIndex()
	rc := newRecent(*cacheSize)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	status := time.NewTicker(5 * time.Second)
	defer status.Stop()

	// Initial sync: ~200 blocks a second up to the tip.
	// After that: one new block a second, like a node that is in sync.
	syncing := time.NewTicker(5 * time.Millisecond)
	defer syncing.Stop()
	following := time.NewTicker(time.Second)
	defer following.Stop()

	var parent [32]byte
	var height uint64
	next := func() {
		raw := fetchBlock(height, parent)
		h := parseHeader(raw)
		if !bytes.Equal(h.Parent, parent[:]) {
			log.Fatalf("block %d: bad parent", height)
		}
		ix.add(h)
		rc.put(height, raw)
		copy(parent[:], h.Hash)
		height++
	}

	log.Printf("syncing to %d", *tip)
	for {
		tick := following.C
		if height < *tip {
			tick = syncing.C
		}
		select {
		case <-tick:
			next()
			if height == *tip {
				log.Printf("synced at %d", height)
			}
		case <-status.C:
			log.Printf("height=%d indexed=%d cached=%d", height, ix.len(), rc.len())
		case <-stop:
			log.Printf("stopping at %d", height)
			return
		}
	}
}
