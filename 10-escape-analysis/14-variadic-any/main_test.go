package main

import "testing"

func TestAllocs(t *testing.T) {
	t.Logf("callCount: %v", testing.AllocsPerRun(1000, func() { callCount(7) }))
	t.Logf("callKeep:  %v", testing.AllocsPerRun(1000, func() { callKeep(7) }))
}
