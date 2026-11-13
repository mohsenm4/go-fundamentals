package main

import (
	"bytes"
	"encoding/binary"
	"sync"
)

// header is the part of a block the index keeps.
type header struct {
	Hash    []byte
	Parent  []byte
	Height  uint64
	Time    int64
	TxCount int
}

func parseHeader(raw []byte) header {
	return header{
		Hash:    bytes.Clone(raw[0:32]),
		Parent:  bytes.Clone(raw[32:64]),
		Height:  binary.BigEndian.Uint64(raw[64:72]),
		Time:    int64(binary.BigEndian.Uint64(raw[72:80])),
		TxCount: int(binary.BigEndian.Uint32(raw[80:84])),
	}
}

// index answers lookups by height and by hash for every block we have seen.
type index struct {
	mu       sync.RWMutex
	byHeight map[uint64]header
	byHash   map[string]uint64
}

func newIndex() *index {
	return &index{
		byHeight: make(map[uint64]header),
		byHash:   make(map[string]uint64),
	}
}

func (ix *index) add(h header) {
	ix.mu.Lock()
	defer ix.mu.Unlock()
	ix.byHeight[h.Height] = h
	ix.byHash[string(h.Hash)] = h.Height
}

func (ix *index) byHeightLookup(height uint64) (header, bool) {
	ix.mu.RLock()
	defer ix.mu.RUnlock()
	h, ok := ix.byHeight[height]
	return h, ok
}

func (ix *index) len() int {
	ix.mu.RLock()
	defer ix.mu.RUnlock()
	return len(ix.byHeight)
}
