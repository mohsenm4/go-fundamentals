package main

import (
	"crypto/sha256"
	"encoding/binary"
	"math/rand/v2"
	"time"
)

// Raw block layout (simplified):
//
//	[0:32]   block hash = sha256 of everything after it
//	[32:64]  parent hash
//	[64:72]  height, big endian
//	[72:80]  timestamp, unix seconds, big endian
//	[80:84]  tx count, big endian
//	[84:]    txs, txSize bytes each
const (
	headerSize = 84
	txSize     = 256
)

// fetchBlock stands in for the network: it returns the raw bytes of the
// block at height, the way a peer would send them.
func fetchBlock(height uint64, parent [32]byte) []byte {
	ntx := 150 + rand.IntN(100)
	raw := make([]byte, headerSize+ntx*txSize)

	copy(raw[32:64], parent[:])
	binary.BigEndian.PutUint64(raw[64:72], height)
	binary.BigEndian.PutUint64(raw[72:80], uint64(time.Now().Unix()))
	binary.BigEndian.PutUint32(raw[80:84], uint32(ntx))
	fillTxs(raw[headerSize:], height)

	sum := sha256.Sum256(raw[32:])
	copy(raw[0:32], sum[:])
	return raw
}

func fillTxs(b []byte, seed uint64) {
	for i := range b {
		b[i] = byte(seed + uint64(i)*31)
	}
}
