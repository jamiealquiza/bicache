package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/jamiealquiza/bicache/v2"
	"github.com/jamiealquiza/tachymeter"
)

func main() {
	concurrency := flag.Int("concurrency", 16, "readers/writers")
	evict := flag.Int("evict", 5, "auto eviction interval (sec)")
	mfu := flag.Int("mfu", 50000, "MFU size")
	mru := flag.Int("mru", 500000, "MRU size")
	ratio := flag.Float64("ratio", 1.02, "Write range size exceeding key space")
	// Add sleep int.
	// Add deleter.
	// Add ttl mode.
	flag.Parse()

	c, _ := bicache.New(&bicache.Config{
		MFUSize:    uint(*mfu),
		MRUSize:    uint(*mru),
		AutoEvict:  uint(*evict * 1000),
		EvictLog:   true,
		ShardCount: 1024,
	})

	keys := int(*ratio * float64((*mfu + *mru)))

	writeT := tachymeter.New(&tachymeter.Config{Size: keys, Safe: true})
	readT := tachymeter.New(&tachymeter.Config{Size: keys * 5, Safe: true})

	for i := 0; i < *concurrency; i++ {
		go readerWriter(c, readT, writeT, keys)
	}

	ticker := time.NewTicker(10 * time.Second)

	for range ticker.C {
		fmt.Printf("\n> Writes:\n")
		fmt.Println(writeT.Calc())

		fmt.Printf("\n> Reads:\n")
		fmt.Println(writeT.Calc())

		stats := c.Stats()
		j, _ := json.Marshal(stats)
		fmt.Printf("\n%s\n", string(j))

		writeT.Reset()
		readT.Reset()
	}
}

func readerWriter(c *bicache.Bicache, rt, wt *tachymeter.Tachymeter, max int) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var start time.Time
	var k string
	var v interface{}
	for {
		k = strconv.Itoa(r.Intn(max))
		time.Sleep(3 * time.Millisecond)
		start = time.Now()
		v = c.Get(k)

		// Write if miss.
		if v == nil {
			start = time.Now()
			c.Set(k, "val")
			wt.AddTime(time.Since(start))
		} else {
			rt.AddTime(time.Since(start))
		}

	}
}
