package fnv

import (
	"strconv"
	"testing"
)

// This needs to be replaced with
// better data.

func BenchmarkHash32a(b *testing.B) {
	d := make([]string, 1024)
	for i := range d {
		d[i] = strconv.Itoa(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Hash32a(d[i%1024])
	}
}

func BenchmarkHash32(b *testing.B) {
	d := make([]string, 1024)
	for i := range d {
		d[i] = strconv.Itoa(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Hash32(d[i%1024])
	}
}

func BenchmarkHash64a(b *testing.B) {
	d := make([]string, 1024)
	for i := range d {
		d[i] = strconv.Itoa(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Hash64a(d[i%1024])
	}
}

func BenchmarkHash64(b *testing.B) {
	d := make([]string, 1024)
	for i := range d {
		d[i] = strconv.Itoa(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Hash64(d[i%1024])
	}
}
