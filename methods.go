// Package bicache implements a two-tier MFU/MRU
// cache with sharded cache units.
package bicache

import (
	"sort"
	"sync/atomic"
	"time"

	"github.com/jamiealquiza/bicache/v2/sll"
	"github.com/jamiealquiza/fnv"
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

// Set takes a key and value and creates
// an entry in the MRU cache. If the key
// already exists, the value is updated.
func (b *Bicache) Set(k string, v interface{}) bool {
	s := b.shards[b.getShard(k)]

	s.Lock()
	ok := s.set(k, v)
	s.Unlock()

	if !ok {
		atomic.AddUint64(&s.counters.overflows, 1)
		return false
	}

	// promoteEvict on write if it's
	// not being handled automatically.
	if !b.autoEvict {
		s.promoteEvict()
	}

	return true
}

// SetTTL is the same as Set but accepts a
// parameter t to specify a TTL in seconds.
func (b *Bicache) SetTTL(k string, v interface{}, t int32) bool {
	s := b.shards[b.getShard(k)]

	s.Lock()

	ok := s.set(k, v)
	if !ok {
		s.Unlock()
		atomic.AddUint64(&s.counters.overflows, 1)
		return false
	}

	// Set the TTL expiration; this is done only after
	// a successful set so that a rejected key doesn't
	// leave an orphaned TTL entry behind. Only count
	// keys that didn't already have a TTL.
	expiration := time.Now().Add(time.Second * time.Duration(t))
	if _, hadTTL := s.ttlMap[k]; !hadTTL {
		atomic.AddUint64(&s.ttlCount, 1)
	}
	s.ttlMap[k] = expiration

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

// set creates or updates a cache entry for key k. New keys
// are created at the MRU head; existing keys have their value
// updated (and are moved to the MRU head if MRU-resident).
// A false is returned if the cache is full and NoOverflow is
// set. The shard lock must be held.
func (s *Shard) set(k string, v interface{}) bool {
	n, exists := s.cacheMap[k]
	if !exists {
		// Reject if we're at capacity
		// and no overflow is set.
		if s.noOverflow && s.mruCache.Len() >= s.mruCap {
			return false
		}

		s.cacheMap[k] = &entry{
			node: s.mruCache.PushHead(&cacheData{k: k, v: v}),
		}

		return true
	}

	n.node.Value.(*cacheData).v = v
	if n.state == 0 {
		s.mruCache.MoveToHead(n.node)
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
		val := read.(*cacheData).v

		s.RUnlock()
		atomic.AddUint64(&s.counters.hits, 1)

		return val
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
		if _, hadTTL := s.ttlMap[k]; hadTTL {
			delete(s.ttlMap, k)
			s.decrementTTLCount(1)
		}
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
// sorted in descending order by score. Returns the
// n top results.
func (b *Bicache) List(n int) ListResults {
	// Make a ListResults large enough to hold the
	// number of cache items present in both cache tiers.
	lr := make(ListResults, 0, b.Size)

	for _, shard := range b.shards {
		shard.RLock()
		for k, v := range shard.cacheMap {
			lr = append(lr, &KeyInfo{
				Key:   k,
				State: v.state,
				Score: atomic.LoadUint64(&v.node.Score),
			})
		}
		shard.RUnlock()
	}

	sort.Sort(lr)
	// return the number
	// of items requested.
	if n < 0 {
		n = 0
	}

	if n < len(lr) {
		return lr[:n]
	}

	return lr
}

// FlushMRU flushes all MRU entries.
func (b *Bicache) FlushMRU() error {
	// Traverse shards.
	for _, s := range b.shards {
		s.Lock()

		// Remove cacheMap entries.
		ttlStart := len(s.ttlMap)
		for k, v := range s.cacheMap {
			if v.state == 0 {
				delete(s.cacheMap, k)
				delete(s.ttlMap, k)
			}
		}
		s.decrementTTLCount(uint64(ttlStart - len(s.ttlMap)))

		s.mruCache = sll.New()

		s.Unlock()
	}

	return nil
}

// FlushMFU flushes all MFU entries.
func (b *Bicache) FlushMFU() error {
	// Traverse shards.
	for _, s := range b.shards {
		s.Lock()

		// Remove cacheMap entries.
		ttlStart := len(s.ttlMap)
		for k, v := range s.cacheMap {
			if v.state == 1 {
				delete(s.cacheMap, k)
				delete(s.ttlMap, k)
			}
		}
		s.decrementTTLCount(uint64(ttlStart - len(s.ttlMap)))

		s.mfuCache = sll.New()

		s.Unlock()
	}

	return nil
}

// FlushAll flushes all cache entries.
// Flush all is much faster than combining both a
// FlushMRU and FlushMFU call.
func (b *Bicache) FlushAll() error {
	// Traverse and reset shard caches.
	for _, s := range b.shards {
		s.Lock()

		// Reset cache and TTL maps and nearest expire.
		s.cacheMap = make(map[string]*entry, s.mfuCap+s.mruCap)
		s.ttlMap = make(map[string]time.Time)
		atomic.StoreUint64(&s.ttlCount, 0)
		s.nearestExpire = time.Now().Add(nearestExpireSentinel)

		// Create new caches.
		s.mfuCache = sll.New()
		s.mruCache = sll.New()

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
// using fnv-1 32 bit based hash-routing
// (we can mask for a modulo since ShardCount
// must be a power of 2).
func (b *Bicache) getShard(k string) int {
	return int(fnv.Hash32(k)) & int(b.ShardCount-1)
}
