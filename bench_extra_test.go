package bicache_test

import (
	"strconv"
	"sync/atomic"
	"testing"

	"github.com/jamiealquiza/bicache/v2"
)

// Worst-case write path: cache at capacity with AutoEvict unset,
// so every Set triggers a synchronous promoteEvict pass.
func BenchmarkSetSyncEvict(b *testing.B) {
	b.StopTimer()

	c, _ := bicache.New(&bicache.Config{
		MFUSize:    10000,
		MRUSize:    30000,
		ShardCount: 512,
	})

	// Fill to capacity.
	for i := 0; i < 40000; i++ {
		c.Set(strconv.Itoa(i), "value")
	}

	keys := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		keys[i] = "k" + strconv.Itoa(i)
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		c.Set(keys[i], "value")
	}
}

func BenchmarkGetParallel(b *testing.B) {
	c, _ := bicache.New(&bicache.Config{
		MFUSize:    10000,
		MRUSize:    600000,
		ShardCount: 1024,
		AutoEvict:  30000,
	})

	const numKeys = 100000
	keys := make([]string, numKeys)
	for i := 0; i < numKeys; i++ {
		keys[i] = strconv.Itoa(i)
		c.Set(keys[i], "value")
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var i int
		for pb.Next() {
			c.Get(keys[i%numKeys])
			i++
		}
	})
}

func BenchmarkSetParallel(b *testing.B) {
	c, _ := bicache.New(&bicache.Config{
		MFUSize:    10000,
		MRUSize:    600000,
		ShardCount: 1024,
		AutoEvict:  30000,
	})

	var ctr uint64

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			n := atomic.AddUint64(&ctr, 1)
			c.Set(strconv.FormatUint(n, 10), "value")
		}
	})
}
