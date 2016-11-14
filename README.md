[![GoDoc](https://godoc.org/github.com/jamiealquiza/bicache?status.svg)](https://godoc.org/github.com/jamiealquiza/bicache)

# bicache
Bicache is a hybrid MFU/MRU cache for Go. It provides a size configurable (in key count [and eventually data size]) two-tier cache, with least-frequently (based on `Get` operations) and least-recently used replacement policies (functionally MFU and MRU, respectively). A shared lookup table is used, limiting read operations to a max of one cache miss even with two tiers of cache. Bicache allows MRU to MFU promotions and overflow evictions at write time or on automatic interval as a background task.

Bicached is designed for roughly p95 single-digit microsecond Sets and 500 nanosecond Gets at 100,000 keys on modern hardware (assuming automatic promotion/eviction is configured and not running; this impact can vary greatly depending on configuration).

# Installation
Requires Go 1.7+.

- `go get -u github.com/jamiealquiza/bicache`
- Import package (or `go install github.com/jamiealquiza/bicache/...` for examples)

# API

See [[GoDoc]](https://godoc.org/github.com/jamiealquiza/bicache) for additional reference.

### Get
`Get(key)` -> returns `value` for `key`

Increments key score by 1.

### Set
`Set(key, value)` -> sets `key` to `value` (if exists, updates)

Moves key to head of MRU cache.

### Delete
`Del(key` -> removes key from cache

# Example

test.go:
```go
package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jamiealquiza/bicache"
)

func main() {
	c := bicache.New(&bicache.Config{
		MfuSize:   5, // MFU capacity in keys
		MruSize:   3, // MRU capacity in keys
		AutoEvict: 500, // Run promotions/evictions automatically every 500ms.
	})

	// Keys and values can be any
	// comparable Go type.
	// Key and value types can be mixed
	// in a single cache object.
	c.Set("name", "john")
	c.Set(1, "a string")
	c.Set(5, 10)
	c.Set("five", 5)

	time.Sleep(time.Second)

	fmt.Println(c.Get("name"))
	fmt.Println(c.Get(1))
	fmt.Println(c.Get(5))
	fmt.Println(c.Get("five"))

	stats := c.Stats()
	j, _ := json.Marshal(stats)
	fmt.Printf("\n%s\n", string(j))
}
```

Output:
```
% go build .
% ./test 
john
a string
10
5

{"MfuSize":1,"MruSize":3,"MfuUsedP":20,"MruUsedP":100}
```

# Misc.
[Diagram](https://raw.githubusercontent.com/jamiealquiza/catpics/master/bicache.jpg)
