package bicache_test

import (
	"log"
	"strconv"
	"testing"
	"time"

	"github.com/jamiealquiza/bicache"
)

// Benchmarks

func BenchmarkGet(b *testing.B) {
	b.StopTimer()

	c, _ := bicache.New(&bicache.Config{
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

	c, _ := bicache.New(&bicache.Config{
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

func BenchmarkSetTtl(b *testing.B) {
	b.StopTimer()

	c, _ := bicache.New(&bicache.Config{
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
		c.SetTtl(keys[i], "my value", 3600)
	}
}

func BenchmarkDel(b *testing.B) {
	b.StopTimer()

	c, _ := bicache.New(&bicache.Config{
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

	for i := 0; i < b.N; i++ {
		c.Set(keys[i], "my value")
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		c.Del(keys[i])
	}
}

func BenchmarkList(b *testing.B) {
	b.StopTimer()

	c, _ := bicache.New(&bicache.Config{
		MfuSize:    10000,
		MruSize:    600000,
		ShardCount: 1024,
		AutoEvict:  30000,
	})

	keys := make([]string, b.N/2)
	for i := 0; i < b.N/2; i++ {
		k := strconv.Itoa(i)
		keys[i] = k
	}

	for i := 0; i < b.N/2; i++ {
		c.Set(keys[i], "my value")
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_ = c.List(b.N / 2)
	}
}

// Tests

func TestSetGet(t *testing.T) {
	c, _ := bicache.New(&bicache.Config{
		MfuSize:    10,
		MruSize:    30,
		ShardCount: 2,
		AutoEvict:  10000,
	})

	ok := c.Set("key", "value")
	if !ok {
		t.Error("Set failed")
	}

	if c.Get("key") != "value" {
		t.Error("Get failed")
	}

	// Ensure that updates work.
	ok = c.Set("key", "value2")
	if !ok {
		t.Error("Update failed")
	}

	if c.Get("key") != "value2" {
		t.Error("Update failed")
	}
}

func TestSetTtl(t *testing.T) {
	c, _ := bicache.New(&bicache.Config{
		MfuSize:    10,
		MruSize:    30,
		ShardCount: 2,
		AutoEvict:  1000,
	})

	ok := c.SetTtl("key", "value", 3)
	if !ok {
		t.Error("Set failed")
	}

	if c.Get("key") != "value" {
		t.Error("Get failed")
	}

	log.Printf("Sleeping for 4 seconds to allow evictions")
	time.Sleep(4 * time.Second)

	if c.Get("key") != nil {
		t.Error("Key TTL expiration failed")
	}
}

func TestDel(t *testing.T) {
	c, _ := bicache.New(&bicache.Config{
		MfuSize:    10,
		MruSize:    30,
		ShardCount: 2,
		AutoEvict:  10000,
	})

	c.Set("key", "value")
	c.Del("key")

	if c.Get("key") != nil {
		t.Error("Delete failed")
	}

	stats := c.Stats()

	if stats.MruSize != 0 {
		t.Errorf("Expected MRU size 0, got %d", stats.MruSize)
	}
}

func TestList(t *testing.T) {
	c, _ := bicache.New(&bicache.Config{
		MfuSize:    10,
		MruSize:    30,
		ShardCount: 2,
		AutoEvict:  1000,
	})

	for i := 0; i < 40; i++ {
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

	log.Printf("Sleeping for 2 seconds to allow evictions")
	time.Sleep(2 * time.Second)

	list := c.List(5)

	if len(list) != 5 {
		t.Errorf("Exected list output len of 5, got %d", len(list))
	}

	// Can only reliably expect the MFU nodes
	// in the list output. Check that the top3
	// are what's expected.
	expected := []string{"2", "0", "1"}
	for i, n := range list[:3] {
		if n.Key != expected[i] {
			t.Errorf(`Expected key "%s" at list element %d, got "%s"`,
				expected[i], i, n.Key)
		}
	}
}

func TestFlushMru(t *testing.T) {
	c, _ := bicache.New(&bicache.Config{
		MfuSize:    10,
		MruSize:    30,
		ShardCount: 1,
		AutoEvict:  1000,
	})

	for i := 0; i < 30; i++ {
		c.Set(strconv.Itoa(i), "value")
	}

	// Check before.
	stats := c.Stats()
	if stats.MruSize != 30 {
		t.Errorf("Expected MFU size of 30, got %d", stats.MfuSize)
	}

	c.FlushMru()

	// Check after.
	stats = c.Stats()
	if stats.MruSize != 0 {
		t.Errorf("Expected MRU size of 0, got %d", stats.MruSize)
	}
}

func TestFlushMfu(t *testing.T) {
	c, _ := bicache.New(&bicache.Config{
		MfuSize:    10,
		MruSize:    30,
		ShardCount: 1,
		AutoEvict:  1000,
	})

	for i := 0; i < 40; i++ {
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

	log.Printf("Sleeping for 2 seconds to allow promotions")
	time.Sleep(2 * time.Second)

	// MFU promotion is already tested in bicache_tests.go
	// TestPromoteEvict. This is somewhat of a dupe, but
	// ensure that if we're testing for a length of 0
	// after a flush that it wasn't 0 to begin with.

	// Check before.
	stats := c.Stats()
	if stats.MfuSize != 3 {
		t.Errorf("Expected MFU size of 3, got %d", stats.MfuSize)
	}

	c.FlushMfu()

	// Check after.
	stats = c.Stats()
	if stats.MfuSize != 0 {
		t.Errorf("Expected MFU size of 0, got %d", stats.MfuSize)
	}
}

func TestFlushAll(t *testing.T) {
	c, _ := bicache.New(&bicache.Config{
		MfuSize:    10,
		MruSize:    30,
		ShardCount: 1,
		AutoEvict:  1000,
	})

	for i := 0; i < 40; i++ {
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

	log.Printf("Sleeping for 2 seconds to allow promotions")
	time.Sleep(2 * time.Second)

	// Check before.
	stats := c.Stats()

	if stats.MfuSize != 3 {
		t.Errorf("Expected MFU size of 3, got %d", stats.MfuSize)
	}

	if stats.MruSize != 30 {
		t.Errorf("Expected MFU size of 30, got %d", stats.MfuSize)
	}

	c.FlushAll()

	// Check after.
	stats = c.Stats()

	if stats.MfuSize != 0 {
		t.Errorf("Expected MFU size of 0, got %d", stats.MfuSize)
	}

	if stats.MruSize != 0 {
		t.Errorf("Expected MRU size of 0, got %d", stats.MruSize)
	}
}
