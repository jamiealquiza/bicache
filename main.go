// The MIT License (MIT)
//
// Copyright (c) 2016 Jamie Alquiza
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.
package bicache

import (
	"container/list"
	"log"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jamiealquiza/bicache/sll"
)

// Bicache implements a two-tier
// cache, combining a MFU and MRU as
// scored linked lists. Each cache key is scored
// by read count in order to track usage frequency.
// New keys are only created in the MRU; the MFU is only
// populated by promoting top-score MRU keys when evictions
// are required. A top-level cache map is used for key lookups
// and routing requests to the appropriate cache.
type Bicache struct {
	sync.RWMutex
	cacheMap      map[interface{}]*entry
	mfuCache      *sll.Sll
	mruCache      *sll.Sll
	mfuCap        uint
	mruCap        uint
	autoEvict     bool
	ttlCount      uint64
	ttlMap        map[interface{}]time.Time
	counters      *Counters
	nearestExpire time.Time
}

// Counters holds Bicache performance
// data.
type Counters struct {
	hits      uint64
	misses    uint64
	evictions uint64
}

// Config holds a Bicache configuration.
// The MFU and MRU cache sizes are set in number
// of keys. The AutoEvict setting specifies an
// interval in milliseconds that a background
// goroutine will handle MRU->MFU promotion
// and MFU/MRU evictions. Setting this to 0
// defers the operation until each Set is called
// on the bicache.
type Config struct {
	MfuSize   uint
	MruSize   uint
	AutoEvict uint
	EvictLog  bool
}

// Entry is a container type for scored
// linked list nodes. Entries are referenced
// in the Bicache cache map and are used to
// locate which cache a lookup should hit.
type entry struct {
	node  *sll.Node
	state uint8 // 0 = MRU, 1 = MFU
}

// cacheData is the data container
// stored in the underlying sll.Node's
// value.
type cacheData struct {
	k interface{}
	v interface{}
}

// Stats holds Bicache
// statistics data.
type Stats struct {
	MfuSize   uint   // Number of acive MFU keys.
	MruSize   uint   // Number of active MRU keys.
	MfuUsedP  uint   // MFU used in percent.
	MruUsedP  uint   // MRU used in percent.
	Hits      uint64 // Cache hits.
	Misses    uint64 // Cache misses.
	Evictions uint64 // Cache evictions.
}

// New takes a *Config and returns
// an initialized *Bicache.
func New(c *Config) *Bicache {
	cache := &Bicache{
		cacheMap:      make(map[interface{}]*entry),
		mfuCache:      sll.New(int(c.MfuSize)),
		mruCache:      sll.New(int(c.MruSize)),
		mfuCap:        c.MfuSize,
		mruCap:        c.MruSize,
		ttlMap:        make(map[interface{}]time.Time),
		counters:      &Counters{},
		nearestExpire: time.Now(),
	}

	var start time.Time
	// Initialize a background goroutine
	// for handling promotions and evictions,
	// if configured.
	if c.AutoEvict > 0 {
		cache.autoEvict = true
		var evicted int
		iter := time.Duration(c.AutoEvict)

		go func(b *Bicache) {
			interval := time.NewTicker(time.Millisecond * iter)
			defer interval.Stop()

			for _ = range interval.C {

				// Run ttl evictions.
				start = time.Now()
				evicted = 0

				// At the very first check, nearestExpire
				// was set to the Bicache initialization time.
				// This is certain to run at least once.
				// The first and real nearest expire will be set
				// in any SetTtl call that's made.
				if b.nearestExpire.Before(start.Add(iter)) {
					evicted = b.EvictTtl()
				}

				if c.EvictLog && evicted > 0 {
					log.Printf("[EvictTtl] %d keys evicted in %s\n", evicted, time.Since(start))
				}

				// Run promotions/overflow evictions.
				start = time.Now()
				b.PromoteEvict()

				if c.EvictLog {
					log.Printf("[AutoEvict] completed in %s\n", time.Since(start))
				}
			}
		}(cache)
	}

	return cache
}

// Stats returns a *Stats with
// Bicache statistics data.
func (b *Bicache) Stats() *Stats {
	b.RLock()
	stats := &Stats{MfuSize: b.mfuCache.Len(), MruSize: b.mruCache.Len()}
	b.RUnlock()

	// Report 0 if MRU-only mode.
	if b.mfuCap > 0 {
		stats.MfuUsedP = uint(float64(stats.MfuSize) / float64(b.mfuCap) * 100)
	} else {
		stats.MfuUsedP = 0
	}

	stats.MruUsedP = uint(float64(stats.MruSize) / float64(b.mruCap) * 100)
	stats.Hits = atomic.LoadUint64(&b.counters.hits)
	stats.Misses = atomic.LoadUint64(&b.counters.misses)
	stats.Evictions = atomic.LoadUint64(&b.counters.evictions)

	return stats
}

// EvictTtl evicts expired keys using a mark
// sweep garbage collection. The number of keys
// evicted is returned.
func (b *Bicache) EvictTtl() int {
	// Return if we have to items with TTLs.
	if atomic.LoadUint64(&b.ttlCount) == 0 {
		return 0
	}

	// Tracking marked expirations
	// in a list.
	expired := list.New()

	// Set initial nearest expire.
	nearestExpire := time.Now().Add(time.Second * 2147483647)

	b.RLock()

	now := time.Now()
	for k, ttl := range b.ttlMap {
		if now.After(ttl) {
			// Add to expired.
			_ = expired.PushBack(k)
		} else {
			// If the key isn't expiring, it is
			// eligible for the nearest expire value.
			if ttl.Before(nearestExpire) {
				nearestExpire = ttl
			}
		}
	}

	b.RUnlock()

	// Lock and evict.
	b.Lock()

	var evicted int
	for k := expired.Front(); k != nil; k = k.Next() {
		if n, exists := b.cacheMap[k.Value]; exists {
			delete(b.cacheMap, k.Value)
			delete(b.ttlMap, k.Value)
			switch n.state {
			case 0:
				b.mruCache.Remove(n.node)
			case 1:
				b.mfuCache.Remove(n.node)
			}
			evicted++
		}
	}

	// Update the ttlCount.
	b.decrementTtlCount(uint64(evicted))

	// Update the nearest expire.
	// If the last TTL'd key was just expired,
	// this will be left at the initially set value
	// at the top of EvictTtl. This means that the
	// auto eviction runs will just skip
	// EvictTtl until a SetTtl creates a real
	// nearest expire timestamp (since it's checking
	// if the nearest expire happens within the auto
	// evict interval).
	b.nearestExpire = nearestExpire

	b.Unlock()

	return evicted
}

// PromoteEvict checks if the MRU exceeds the
// Config.MruSize (overflow count) If so, the top <overflow count>
// MRU scores are checked against the MFU. If any of the top MRU scores
// are greater than the lowest MFU scores, they are promoted
// to the MFU (if possible). Any remaining overflow count
// is evicted from the tail of the MRU.
func (b *Bicache) PromoteEvict() {
	//b.mfuCache.PreSort() // This may require a lock in sll that
	// would conflict with with a Get score increment.

	b.Lock()
	defer b.Unlock()

	// How far over MRU capacity are we?
	mruOverflow := int(b.mruCache.Len() - b.mruCap)
	if mruOverflow <= 0 {
		return
	}

	// If MFU cap is 0, shortcut to
	// LRU-only behavior.
	if b.mfuCap == 0 {
		b.evictFromMruTail(mruOverflow)
		return
	}

	// Get the top n MRU elements
	// where n = MRU capacity overflow.
	mruToPromoteEvict := b.mruCache.HighScores(mruOverflow)
	// Put into ascending order.
	sort.Sort(sort.Reverse(mruToPromoteEvict))

	// Check MFU capacity.
	mfuFree := int(b.mfuCap - b.mfuCache.Len())
	if mfuFree < 0 {
		mfuFree = 0
	}

	// canPromote is the count of mruOverflow
	// that can fit into currently unused MFU slots.
	// This is only likely to be met if this
	// is a somewhat new cache.
	var canPromote int
	if int(mfuFree) >= mruOverflow {
		canPromote = mruOverflow
	} else {
		canPromote = mfuFree
	}

	// If the MFU is already full,
	// we can skip the next block.
	if canPromote == 0 {
		goto promoteByScore
	}

	// This is all MRU->MFU promotion
	// using free slots.
	if canPromote > 0 {
		for _, node := range mruToPromoteEvict[:canPromote] {
			// Create a new node at the tail of the MFU,
			// copy values, update cacheMap state/reference.
			newNode := b.mfuCache.PushTail(node.Value)
			newNode.Score = node.Score
			// The original key is stored in the cacheData.key
			// field as the node's Value. This allows a reverse
			// lookup of a cacheMap entry by node.
			b.cacheMap[node.Value.(*cacheData).k].state = 1
			b.cacheMap[node.Value.(*cacheData).k].node = newNode
			b.mruCache.Remove(node)
		}

		// If we were able to promote
		// all the overflow, return.
		if canPromote == mruOverflow {
			return
		}
	}

promoteByScore:

	// Get a remainder to either promote by score
	// to the MFU or ultimately evict from the MRU.
	mruOverflow -= canPromote
	remainderPosition := canPromote

	// Counter to track
	// how many from the MRU keys
	// were promoted by score.
	var promotedByScore int

	// We're here on two conditions:
	// 1) The MFU was full. We need to handle all mruToPromoteEvict (canPromote == 0).
	// 2) We promoted some mruToPromoteEvict and have leftovers (canPromote > 0).

	// Get top MRU scores and bottom MFU scores to compare.
	bottomMfu := b.mfuCache.LowScores(mruOverflow)

	// If the lowest MFU score is higher than the lowest
	// score to promote, none of these are eligible.
	// Promoting by score is an expensive operation,
	// it's desireable to skip if possible.
	// TODO it may be possible that a batch of many items
	// from the MRU overflow is elgible by score, where the
	// rest would fail. Need to add an additional short circuit
	// assuming we enter the promote by score routine.
	if bottomMfu[0].Score >= mruToPromoteEvict[remainderPosition].Score {
		goto evictFromMruTail
	}

	// Otherwise, scan for a replacement.
	for _, mruNode := range mruToPromoteEvict[remainderPosition:] {
		for i, mfuNode := range bottomMfu {
			if mruNode.Score > mfuNode.Score {
				// Create a new node at the MRU head,
				// then copy the evicted MFU node over.
				newMruNode := b.mruCache.PushHead(mfuNode.Value)
				newMruNode.Score = mfuNode.Score
				b.cacheMap[mfuNode.Value.(*cacheData).k].state = 0
				b.cacheMap[mfuNode.Value.(*cacheData).k].node = newMruNode
				b.mfuCache.Remove(mfuNode)

				// Promote the MRU node to the MFU.
				newMfuNode := b.mfuCache.PushTail(mruNode.Value)
				newMfuNode.Score = mruNode.Score
				b.cacheMap[mruNode.Value.(*cacheData).k].state = 1
				b.cacheMap[mruNode.Value.(*cacheData).k].node = newMfuNode
				b.mruCache.Remove(mruNode)

				promotedByScore++

				// Remove the replaced MFU node from the
				// bottomMfu list so it's not attempted twice.
				bottomMfu = append(bottomMfu[:i], bottomMfu[i+1:]...)
				break
			}
		}

	}

evictFromMruTail:
	// What's the overflow remainder count?
	toEvict := mruOverflow - promotedByScore
	// Evict this many from the MRU tail.
	b.evictFromMruTail(toEvict)

}

func (b *Bicache) evictFromMruTail(n int) {
	for i := 0; i < n; i++ {
		node := b.mruCache.Tail()
		delete(b.cacheMap, node.Value.(*cacheData).k)
		b.mruCache.RemoveTail()

		var evicted int

		// Check if this key existed in the
		// ttl map. Clean up entry / counter, if so.
		if _, exists := b.ttlMap[node.Value.(*cacheData).k]; exists {
			delete(b.ttlMap, node.Value.(*cacheData).k)
			evicted++
		}

		// Update the ttlCount.
		b.decrementTtlCount(uint64(evicted))
	}
}

// decrementTtlCount decrements the Bicache.ttlCount
// value by n. This should only be performed within
// a locked mutex.
func (b *Bicache) decrementTtlCount(n uint64) {
	// Prevents some obscure
	// scenario where ttlCount is
	// already 0 and we rollover to
	// uint max.
	if b.ttlCount-n > b.ttlCount {
		atomic.StoreUint64(&b.ttlCount, 0)
	} else {
		atomic.StoreUint64(&b.ttlCount, b.ttlCount-n)
	}

	// Increment the evictions count
	// by n, regardless.
	atomic.AddUint64(&b.counters.evictions, n)
}
