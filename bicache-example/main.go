package main

import (
	"fmt"

	"github.com/jamiealquiza/bicache"
)

func main() {
	c := bicache.New(&bicache.Config{
		MfuSize: 10,
		MruSize: 100,
		Safe:    true,
	})

	c.Set("key", "val")
	fmt.Println(c.Get("key"))
}
