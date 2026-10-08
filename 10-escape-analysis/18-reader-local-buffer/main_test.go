package main

import (
	"io"
	"strings"
	"testing"
)

func TestAllocs(t *testing.T) {
	r := strings.NewReader("hello")
	var ri io.Reader = r
	t.Logf("readIface:    %v", testing.AllocsPerRun(1000, func() { readIface(ri) }))
	t.Logf("readConcrete: %v", testing.AllocsPerRun(1000, func() { readConcrete(r) }))
}

func BenchmarkIface(b *testing.B) {
	r := strings.NewReader("hello")
	var ri io.Reader = r
	for b.Loop() {
		r.Reset("hello")
		readIface(ri)
	}
}

func BenchmarkConcrete(b *testing.B) {
	r := strings.NewReader("hello")
	for b.Loop() {
		r.Reset("hello")
		readConcrete(r)
	}
}
