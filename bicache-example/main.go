package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jamiealquiza/bicache"
	"github.com/jamiealquiza/tachymeter"
)

func main() {
	c := bicache.New(&bicache.Config{
		MfuSize:   500,
		MruSize:   500,
		AutoEvict: 1000,
	})

	keys := 10000

	t := tachymeter.New(&tachymeter.Config{Size: keys})
	fmt.Printf("[ Set %d keys ]\n", keys)
	for i := 0; i < keys; i++ {
		start := time.Now()
		c.Set(i, i)
		t.AddTime(time.Since(start))
	}
	t.AddCount(keys)
	t.Calc().Dump()

	fmt.Println()

	c.Get(3)
	c.Get(3)
	c.Get(3)
	c.Get(3)
	c.Get(3)

	//time.Sleep(1 * time.Second)

	c.Get(2)
	c.Get(2)
	c.Get(2)
	c.Get(2)
	c.Get(2)

	time.Sleep(3 * time.Second)

	t.Reset()
	fmt.Printf("[ Get %d keys ]\n", keys)
	for i := 0; i < keys; i++ {
		start := time.Now()
		_ = c.Get(i)
		t.AddTime(time.Since(start))
	}
	t.AddCount(keys)

	t.Calc().Dump()

	stats := c.Stats()
	j, _ := json.Marshal(stats)
	fmt.Printf("\n%s\n", string(j))
}
