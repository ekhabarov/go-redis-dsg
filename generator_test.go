package main

import "testing"

func BenchmarkGeneration(b *testing.B) {
	for i := 0; i < b.N; i++ {
		createMsg()
	}
}
