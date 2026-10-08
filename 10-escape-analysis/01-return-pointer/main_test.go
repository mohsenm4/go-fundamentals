package main

import "testing"

func BenchmarkPointer(b *testing.B) {
	for b.Loop() {
		createUser()
	}
}

func BenchmarkValue(b *testing.B) {
	for b.Loop() {
		createValue()
	}
}
