package main

import (
	"fmt"
	"encoding/json"

	"github.com/jamiealquiza/bicache"
)

func main() {
	c := bicache.New(&bicache.Config{
		MfuSize: 10,
		MruSize: 100,
	})

	c.Set("key", "val")
	fmt.Println(c.Get("key"))

	stats := c.Stats()
	j, _ := json.Marshal(stats)
	fmt.Println(string(j))

	c.Delete("key")
	fmt.Println(c.Get("key"))
}