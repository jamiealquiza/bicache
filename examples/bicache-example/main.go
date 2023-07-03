package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/jamiealquiza/bicache/v2"
	"github.com/jamiealquiza/tachymeter"
)

func main() {
	c, _ := bicache.New(&bicache.Config{
		MFUSize:    50000,
		MRUSize:    50000,
		AutoEvict:  1000,
		ShardCount: 512,
	})

	keys := 100000

	// Cache pre-warm.
	for i := 0; i < keys; i++ {
		key := strconv.Itoa(i)
		c.Set(key, []byte{0})
	}

	t := tachymeter.New(&tachymeter.Config{Size: keys})
	fmt.Printf("[ Set %d keys ]\n", keys)
	for i := 0; i < keys; i++ {
		key := strconv.Itoa(i)
		start := time.Now()
		c.Set(key, []byte{0})
		t.AddTime(time.Since(start))
	}
	fmt.Println(t.Calc())

	fmt.Println()

	t.Reset()
	fmt.Printf("[ Get %d keys ]\n", keys)
	for i := 0; i < keys; i++ {
		key := strconv.Itoa(i)
		start := time.Now()
		_ = c.Get(key)
		t.AddTime(time.Since(start))
	}

	fmt.Println(t.Calc())

	stats := c.Stats()
	j, _ := json.Marshal(stats)
	fmt.Printf("\n%s\n", string(j))
}
