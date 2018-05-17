package bicache_test

import (
	"log"
	"strconv"
	"testing"
	"time"

	"github.com/jamiealquiza/bicache"
)

func TestNew(t *testing.T) {
	// MFU size, MRU size, shard count, expected bicache size.
	configExpected := [][4]int{
		[4]int{0, 1, 512, 512},
		[4]int{1, 2, 512, 1024},
		[4]int{256, 1024, 512, 1536},
		[4]int{10000, 30000, 512, 40448}}

	// Check that the cache sizes are fitted to the
	// minimum for the number of shards specified.
	for _, n := range configExpected {
		c, _ := bicache.New(&bicache.Config{
			MFUSize:    uint(n[0]),
			MRUSize:    uint(n[1]),
			ShardCount: n[2],
		})

		if c.Size != n[3] {
			t.Errorf("Expected bicache size %d, got %d", n[3], c.Size)
		}
	}
}

func TestStats(t *testing.T) {
	c, _ := bicache.New(&bicache.Config{
		MFUSize:    10,
		MRUSize:    30,
		ShardCount: 2,
		AutoEvict:  3000,
	})

	for i := 0; i < 50; i++ {
		c.Set(strconv.Itoa(i), "value")
		c.Get(strconv.Itoa(i))
		c.Get(strconv.Itoa(i))
	}

	c.Get("nil")

	log.Printf("Sleeping for 4 seconds to allow evictions")
	time.Sleep(4 * time.Second)

	stats := c.Stats()

	if stats.MFUSize != 10 {
		t.Errorf("Expected MFU size 10, got %d", stats.MFUSize)
	}

	if stats.MRUSize != 30 {
		t.Errorf("Expected MRU size 30, got %d", stats.MRUSize)
	}

	if stats.MFUUsedP != 100 {
		t.Errorf("Expected MFU usedp 100, got %d", stats.MFUUsedP)
	}

	if stats.MRUUsedP != 100 {
		t.Errorf("Expected MRU usedp 100, got %d", stats.MRUUsedP)
	}

	if stats.Hits != 100 {
		t.Errorf("Expected 100 hits, got %d", stats.Hits)
	}

	if stats.Misses != 1 {
		t.Errorf("Expected 1 misses, got %d", stats.Misses)
	}

	if stats.Evictions != 10 {
		t.Errorf("Expected 10 evictions, got %d", stats.Evictions)
	}

	if stats.Overflows != 0 {
		t.Errorf("Expected 0 overflows, got %d", stats.Overflows)
	}
}

func TestEvictTtl(t *testing.T) {
	c, _ := bicache.New(&bicache.Config{
		MFUSize:    10,
		MRUSize:    30,
		ShardCount: 2,
		AutoEvict:  1000,
	})

	c.SetTTL("5", "value", 5)
	c.SetTTL("30", "value", 30)

	if v := c.Get("5"); v != "value" {
		t.Error("Expected hit")
	}

	log.Printf("Sleeping for 6 seconds to allow evictions")
	time.Sleep(6 * time.Second)

	// Check that the value was evicted.
	if v := c.Get("5"); v != nil {
		t.Error("Expected miss")
	}

	stats := c.Stats()

	if stats.MRUSize != 1 || stats.Evictions != 1 {
		t.Error("Unexpected stats")
	}
}

func TestPromoteEvict(t *testing.T) {
	// Also covers MRU tail eviction.
	c, _ := bicache.New(&bicache.Config{
		MFUSize:    10,
		MRUSize:    30,
		ShardCount: 2,
		AutoEvict:  5000,
	})

	for i := 0; i < 50; i++ {
		c.Set(strconv.Itoa(i), "value")
	}

	c.Get("0")
	c.Get("0")
	c.Get("0")
	c.Get("0")

	c.Get("1")
	c.Get("1")
	c.Get("1")

	c.Get("2")
	c.Get("2")
	c.Get("2")
	c.Get("2")
	c.Get("2")
	c.Get("2")

	log.Printf("Sleeping for 6 seconds to allow evictions")
	time.Sleep(6 * time.Second)

	stats := c.Stats()

	if stats.MFUSize != 3 {
		t.Error("Unexpected MFU count")
	}

	if stats.MRUSize != 30 {
		t.Error("Unexpected MRU count")
	}

	// Check that the expected MRU members exist.
	for k := 20; k <= 49; k++ {
		if v := c.Get(strconv.Itoa(k)); v == nil {
			t.Errorf("Unexpected miss for key %d", k)
		}
	}

	// Check that the expected MFU members exist.
	for _, k := range []int{0, 1, 2} {
		if v := c.Get(strconv.Itoa(k)); v == nil {
			t.Errorf("Unexpected miss for key %d", k)
		}
	}
}
