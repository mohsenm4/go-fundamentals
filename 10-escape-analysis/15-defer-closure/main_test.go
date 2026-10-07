package main

import "testing"

func TestAllocs(t *testing.T) {
	t.Logf("once:   %v", testing.AllocsPerRun(1000, func() { once() }))
	t.Logf("inLoop: %v", testing.AllocsPerRun(1000, func() { inLoop() }))
}
