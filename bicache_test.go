package bicache

import (
	"strconv"
	"testing"
)

func BenchmarkGet(b *testing.B) {
	b.StopTimer()

	c := New(&Config{
		MfuSize:    10000,
		MruSize:    600000,
		ShardCount: 1024,
		AutoEvict:  30000,
	})

	keys := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		k := strconv.Itoa(i)
		keys[i] = k
		c.Set(k, "my value")
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		c.Get(keys[i])
	}
}

func BenchmarkSet(b *testing.B) {
	b.StopTimer()

	c := New(&Config{
		MfuSize:    10000,
		MruSize:    600000,
		ShardCount: 1024,
		AutoEvict:  30000,
	})

	keys := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		k := strconv.Itoa(i)
		keys[i] = k
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		c.Set(keys[i], "my value")
	}
}
