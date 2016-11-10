package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jamiealquiza/tachymeter"
	"github.com/jamiealquiza/bicache"
)

func main() {
	c := bicache.New(&bicache.Config{
		MfuSize: 50000,
		MruSize: 50000,
		AutoEvict: 5000,
	})

	t := tachymeter.New(&tachymeter.Config{Size: 100000})
	fmt.Println("[ Set 100000 keys ]")
	for i := 0; i < 100000; i++ {
		start := time.Now()
		c.Set(i, i*2)
		t.AddTime(time.Since(start))
	}
	t.AddCount(10000)
	t.Calc().Dump()

	fmt.Println()

	t.Reset()
	fmt.Println("[ Get 100000 keys ]")
	for i := 0; i < 100000; i++ {
		start := time.Now()
		_ = c.Get(i)
		t.AddTime(time.Since(start))
	}
	t.AddCount(10000)

	t.Calc().Dump()

	stats := c.Stats()
	j, _ := json.Marshal(stats)
	fmt.Println(string(j))
}
