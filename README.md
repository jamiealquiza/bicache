[![GoDoc](https://godoc.org/github.com/jamiealquiza/bicache?status.svg)](https://godoc.org/github.com/jamiealquiza/bicache)

# bicache
Bicache is a hybrid MFU/MRU, TTL-optional, general key cache for Go. Why should you be interested? Pure MRU caches are great because they're fast, simple and safe. Items that are used often generally remain in the cache. The problem is that a single, large scan where the number of misses exceeds the MRU cache size causes _highly used_ (and perhaps the most useful) data to be evicted in favor of _recent_ data. A MFU cache makes the distinction of item value vs recency based on a history of read counts. This means that valuable keys are insulated from large scans of potentially less valuable data.

Bicache's two tiers of cache are individually size configurable (in key count [and eventually data size]). A shared lookup table is used, limiting read ops to a max of one cache miss over two tiers of cache. Bicache allows MRU to MFU promotions and overflow evictions at write time or on automatic interval as a background task.

Bicached averages roughly p95 single-digit microsecond Sets and 500 nanosecond Gets at 100,000 keys on modern hardware (assuming a promotion/eviction is not running; the impact can vary greatly depending on configuration). This translates to millions of get/set operations per second from a single thread.

# Diagram

In a MRU cache, both fetching and setting a key moves it to the front of the list. When the list is full, keys are evicted from the tail when space for a new key is needed. An item that is hit often (in orange) remains in the cache by probability that it was accessed recently.

Commonly, a cache miss works as follows: `cache get` -> `miss` -> `backend lookup` -> `cache results`. If a piece of software were to traverse a large list of user IDs stored in a backend database, it's likely that the cache capacity will be much smaller than the number of user IDs available in the database. This will result in the entire MRU being flushed and replaced.
![img](https://raw.githubusercontent.com/jamiealquiza/catpics/master/mru.png)

Bicache isolates MRU large scan evictions by promoting the most used keys to an MFU cache when the MRU cache is full. MRU to MFU promotions are intelligent; rather than attempting to promote tail items, Bicache asynchronously gathers the highest score MRU keys and promotes those that have scores exceeding keys in the MFU. Any remainder key count that must be evicted relegates to MFU to MRU demotion followed by MRU tail eviction.

New keys are always set to the head of the MRU list; MFU keys are only ever set by promotion. 
![img](https://raw.githubusercontent.com/jamiealquiza/catpics/master/mfu-mru.png)



# Installation
Test with Go 1.7.

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

### SetTtl
`Set(key, value, ttl)` -> sets `key` to `value` (if exists, updates) with a `ttl` expiration (in seconds)

Moves key to head of MRU cache.

### Del
`Del(key)` -> removes key from cache

### List
`List(int)` -> returns the top n key names:state:score (state: 0 = MRU cache, 1 = MFU cache)

# Configuration

### Cache sizes

Bicache can be configured with arbitrary sizes for each cache, allowing a ratio of MFU to MRU for different usage patterns. While the example shows very low cache sizes, this is purely to demonstrate functionality when the MRU is overflowed. A real world configuration might be a 10,000 key MFU and 30,000 key MRU capacity.

The MFU can also be set to 0, causing Bicache to behave like a typical MRU/LRU cache.

### Auto Eviction

TTL expirations, MRU to MFU promotions, and MRU overflow evictions only occur automatically if the `AutoEvict` configuration parameter is set. This is a background task that only runs if a non-zero parameter is set. If unset or explicitely configured to 0, TTL expirations never run and MRU promotions and evictions will be performed at each Set operation.

The Bicache `EvictLog` configuratio parameter specifies whether or not eviction timing logs are emitted.

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
		AutoEvict: 500, // Run TTL evictions + MRU->MFU promotions / evictions automatically every 500ms.
		EvictLog: true, // Emit eviction timing logs.
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
