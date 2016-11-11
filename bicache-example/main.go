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
		MfuSize:   50000,
		MruSize:   25000,
		AutoEvict: 1000,
	})

	t := tachymeter.New(&tachymeter.Config{Size: 50000})
	fmt.Println("[ Set 100000 keys ]")
	for i := 0; i < 50000; i++ {
		start := time.Now()
		c.Set(i, i*2)
		t.AddTime(time.Since(start))
	}
	t.AddCount(50000)
	t.Calc().Dump()

	fmt.Println()
	time.Sleep(3*time.Second)

	t.Reset()
	fmt.Println("[ Get 100000 keys ]")
	for i := 0; i < 50000; i++ {
		start := time.Now()
		_ = c.Get(i)
		t.AddTime(time.Since(start))
	}
	t.AddCount(50000)

	t.Calc().Dump()

	stats := c.Stats()
	j, _ := json.Marshal(stats)
	fmt.Printf("\n%s\n", string(j))
}
