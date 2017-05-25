// Package bicache implements a two-tier MFU/MRU
// cache with sharded cache units.
package bicache

import (
	"sort"
	"sync/atomic"
	"time"

	"github.com/jamiealquiza/bicache/sll"
)

// KeyInfo holds a key name, state (0: MRU, 1: MFU)
// and cache score.
type KeyInfo struct {
	Key   string
	State uint8
	Score uint64
}

// ListResults is a container that holds results from
// from a List method (as keyInfo), allowing sorting of
// available key names by score.
type ListResults []*KeyInfo

// listResults methods to satisfy the sort interface.

func (lr ListResults) Len() int {
	return len(lr)
}

func (lr ListResults) Less(i, j int) bool {
	// Note operator set for desc. order.
	return lr[i].Score > lr[j].Score
}

func (lr ListResults) Swap(i, j int) {
	lr[i], lr[j] = lr[j], lr[i]
}

// Bicache is storing a [2]interface{}
// as the underlying sll node's value.
// Position 0 is the node's key and position
// 1 is the value. This is done so that
// the node a node can be looked up in the
// cache map if the key would otherwise be
// unknown.

// Set takes a key and value and creates
// and entry in the MRU cache. If the key
// already exists, the value is updated.
func (b *Bicache) Set(k string, v interface{}) bool {
	s := b.shards[b.getShard(k)]

	s.Lock()
	// If the entry exists, update. If not,
	// create at the tail of the MRU cache.
	if n, exists := s.cacheMap[k]; !exists {
		// Return false if we're at capacity
		// and no overflow is set.
		if s.noOverflow && s.mruCache.Len() >= s.mruCap {
			s.Unlock()
			atomic.AddUint64(&s.counters.overflows, 1)
			return false
		}

		// Create at the MRU tail.
		s.cacheMap[k] = &entry{
			node: s.mruCache.PushHead(&cacheData{k: k, v: v}),
		}
	} else {
		n.node.Value.(*cacheData).v = v
		if n.state == 0 {
			s.mruCache.MoveToHead(n.node)
		}
	}

	s.Unlock()

	// promoteEvict on write if it's
	// not being handled automatically.
	if !b.autoEvict {
		s.promoteEvict()
	}

	return true
}

// SetTTL is the same as set but accepts a
// parameter t to specify a TTL in seconds.
func (b *Bicache) SetTTL(k string, v interface{}, t int32) bool {
	s := b.shards[b.getShard(k)]

	s.Lock()

	// Set TTL expiration
	expiration := time.Now().Add(time.Second * time.Duration(t))
	s.ttlMap[k] = expiration

	// Increment TTL counter.
	atomic.AddUint64(&s.ttlCount, 1)

	// Proceed to normal Set operation.
	// This logic is duplicated for now
	// to skip releasing / re-acquiring a mutex.

	// If the entry exists, update. If not,
	// create at the tail of the MRU cache.
	if n, exists := s.cacheMap[k]; !exists {
		// Return false if we're at capacity
		// and no overflow is set.
		if s.noOverflow && s.mruCache.Len() >= s.mruCap {
			s.Unlock()
			atomic.AddUint64(&s.counters.overflows, 1)
			return false
		}
		// Create at the MRU tail.
		s.cacheMap[k] = &entry{
			node: s.mruCache.PushHead(&cacheData{k: k, v: v}),
		}
	} else {
		n.node.Value.(*cacheData).v = v
		if n.state == 0 {
			s.mruCache.MoveToHead(n.node)
		}
	}

	// Update the nearest expire.
	if expiration.Before(s.nearestExpire) {
		s.nearestExpire = expiration
	}

	s.Unlock()

	// promoteEvict on write if it's
	// not being handled automatically.
	if !b.autoEvict {
		s.promoteEvict()
	}

	return true
}

// Get takes a key and returns the value. Every get
// on a key increases the key score.
func (b *Bicache) Get(k string) interface{} {
	s := b.shards[b.getShard(k)]

	s.RLock()

	if n, exists := s.cacheMap[k]; exists {
		read := n.node.Read()

		s.RUnlock()
		atomic.AddUint64(&s.counters.hits, 1)

		return read.(*cacheData).v
	}

	s.RUnlock()
	atomic.AddUint64(&s.counters.misses, 1)

	return nil
}

// Del deletes a key.
func (b *Bicache) Del(k string) {
	s := b.shards[b.getShard(k)]

	s.Lock()

	if n, exists := s.cacheMap[k]; exists {
		delete(s.cacheMap, k)
		delete(s.ttlMap, k)
		switch n.state {
		case 0:
			s.mruCache.Remove(n.node)
		case 1:
			s.mfuCache.Remove(n.node)
		}
	}

	s.Unlock()
}

// List returns all key names, states, and scores
// sorted in descending order by score. Returns n
// top restults.
func (b *Bicache) List(n int) ListResults {
	// Make a ListResults large enough to hold the
	// number of cache items present in both cache tiers.
	lr := make(ListResults, 0, b.Size)

	var i int
	for _, shard := range b.shards {
		for k, v := range shard.cacheMap {
			lr = append(lr, &KeyInfo{
				Key:   k,
				State: v.state,
				Score: v.node.Score,
			})
			i++
		}
	}

	sort.Sort(lr)
	// return the number
	// of items requested.
	if n < len(lr) {
		return lr[:n]
	}

	return lr
}

// FlushMru flushes all MRU entries.
func (b *Bicache) FlushMru() error {
	// Traverse shards.
	for _, s := range b.shards {
		s.Lock()

		// Remove cacheMap entries.
		for k, v := range s.cacheMap {
			if v.state == 0 {
				delete(s.cacheMap, k)
				delete(s.ttlMap, k)
			}
		}

		s.mruCache = sll.New(int(s.mruCap))

		s.Unlock()
	}

	return nil
}

// FlushMfu flushes all MFU entries.
func (b *Bicache) FlushMfu() error {
	// Traverse shards.
	for _, s := range b.shards {
		s.Lock()

		// Remove cacheMap entries.
		for k, v := range s.cacheMap {
			if v.state == 1 {
				delete(s.cacheMap, k)
				delete(s.ttlMap, k)
			}
		}

		s.mfuCache = sll.New(int(s.mfuCap))

		s.Unlock()
	}

	return nil
}

// FlushAll flushes all cache entries.
// Flush all is much faster than combining both a
// FlushMru and FlushMfu call.
func (b *Bicache) FlushAll() error {
	// Traverse and reset shard caches.
	for _, s := range b.shards {
		s.Lock()

		// Reset cache and TTL maps and nearest expire.
		s.cacheMap = make(map[string]*entry, s.mfuCap+s.mruCap)
		s.ttlMap = make(map[string]time.Time)
		s.nearestExpire = time.Now().Add(time.Second * 2147483647)

		// Create new caches.
		s.mfuCache = sll.New(int(s.mfuCap))
		s.mruCache = sll.New(int(s.mruCap))

		s.Unlock()
	}

	return nil
}

// Pause suspends normal and TTL evictions.
// If eviction logging is enabled, bicache
// will log that evictions are paused
// at each interval if paused.
func (b *Bicache) Pause() error {
	atomic.StoreUint32(&b.paused, 1)
	return nil
}

// Resume resumes normal and TTL evictions.
func (b *Bicache) Resume() error {
	atomic.StoreUint32(&b.paused, 0)
	return nil
}

// getShard returns the shard index
// using fnv-1 32-bit based hash-routing.
func (b *Bicache) getShard(k string) int {
	var h = 2166136261
	for _, c := range []byte(k) {
		h *= 16777619
		h ^= int(c)
	}

	return h & int(b.ShardCount-1)
}
