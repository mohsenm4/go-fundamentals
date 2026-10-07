package main

import "testing"

func TestAllocs(t *testing.T) {
	t.Logf("localOnly: %v", testing.AllocsPerRun(1000, func() { localOnly() }))
	t.Logf("withPut:   %v", testing.AllocsPerRun(1000, func() { withPut() }))
}
