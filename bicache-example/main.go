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
		MfuSize: 100,
		MruSize: 100000,
	})

	c.Set("one", "one")
	c.Set("two", "two")
	c.Set("three", "three")
	c.Set("four", "four")
	c.Set("five", "five")
	c.Set("six", "six")

	fmt.Println(c.Get("one"))

	fmt.Println()

	stats := c.Stats()
	j, _ := json.Marshal(stats)
	fmt.Println(string(j))

	c.Delete("one")

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
}
