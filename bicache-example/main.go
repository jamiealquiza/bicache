package main

import (
	"encoding/json"
	"fmt"

	"github.com/jamiealquiza/bicache"
)

func main() {
	c := bicache.New(&bicache.Config{
		MfuSize: 10,
		MruSize: 3,
	})

	c.Set("one", "one")
	c.Set("two", "two")
	c.Set("three", "three")
	c.Set("four", "four")
	c.Set("five", "five")
	c.Set("six", "six")

	fmt.Println(c.Get("one"))

	stats := c.Stats()
	j, _ := json.Marshal(stats)
	fmt.Println(string(j))

	c.Delete("one")
	fmt.Println(c.Get("one"))
}
