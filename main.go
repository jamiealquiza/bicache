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
	"sort"
	"sync"
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
	cacheMap  map[interface{}]*entry
	mfuCache  *sll.Sll
	mruCache  *sll.Sll
	mfuCap    uint
	mruCap    uint
	autoEvict bool
	// MFU top/bottom scores.
}

// Config holds a Bicache configuration.
// The MFU and MRU cache sizes are set in number
// of keys.
type Config struct {
	MfuSize   uint
	MruSize   uint
	AutoEvict uint // on-write vs automatic
}

// Entry is a container type for scored
// linked list nodes. Entries are referenced
// in the Bicache cache map and are used to
// locate which cache a lookup should hit.
type entry struct {
	node  *sll.Node
	state uint8 // 0 = MRU, 1 = MFU
}

// Stats holds Bicache
// statistics data.
type Stats struct {
	MfuSize  uint // Number of acive MFU keys.
	MruSize  uint // Number of active MRU keys.
	MfuUsedP uint // MFU used in percent.
	MruUsedP uint // MRU used in percent.
}

// New takes a *Config and returns
// an initialized *Bicache.
func New(c *Config) *Bicache {
	cache := &Bicache{
		cacheMap: make(map[interface{}]*entry),
		mfuCache: sll.New(),
		mruCache: sll.New(),
		mfuCap:   c.MfuSize,
		mruCap:   c.MruSize,
	}

	if c.AutoEvict > 0 {
		cache.autoEvict = true
		go func(b *Bicache) {
			interval := time.NewTicker(time.Millisecond * time.Duration(c.AutoEvict))
			defer interval.Stop()

			for _ = range interval.C {
				b.PromoteEvict()
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

	stats.MfuUsedP = uint(float64(stats.MfuSize) / float64(b.mfuCap) * 100)
	stats.MruUsedP = uint(float64(stats.MruSize) / float64(b.mruCap) * 100)

	return stats
}

// PromoteEvict checks if the MRU exceeds the
// Config.MruSize. If so, the top MRU scores are
// checked against the MFU. If any of the top MRU scores
// are greater than the lowest MFU scores, they are promoted
// to the MFU (if possible). Any remaining count of evictions
// that must occur are removed from the tail of the MRU.
func (b *Bicache) PromoteEvict() {
	b.Lock()
	defer b.Unlock()

	// How far over MRU capacity are we?
	mruOverflow := int(b.mruCache.Len() - b.mruCap)
	if mruOverflow <= 0 {
		return
	}

	// Get the top n MRU elements
	// where n = MRU capacity overflow.
	topMru := b.mruCache.HighScores(mruOverflow)
	// Put into ascending order.
	sort.Sort(sort.Reverse(topMru))

	// Check MFU capacity.
	mfuFree := b.mfuCap - b.mfuCache.Len()

	// Promote what we can.
	// Can promote is the count of mruOverflow
	// that can fit into currently unused MFU slots.
	canPromote := int(mfuFree) - (int(mfuFree) - mruOverflow)
	if canPromote > 0 {
		for _, node := range topMru[:canPromote] {
			// Need to update the state.
			b.cacheMap[node.Value.([2]interface{})[0]].state = 1
			// Remove from MRU.
			b.mruCache.Remove(node)
			// Move to MFU.
			b.mfuCache.PushTailNode(node)
		}

	}

	// Get count of overflow that coulnd't be
	// freely promoted.
	// 1) Check if it can be promoted with
	// a greater score.
	// 2) Evict remainder from MRU tail.
	/*
		if canPromote < mruOverflow {
			remainder := promoteByScore(topMru[canPromote:])
		} /*else {
			for _, node := range
				delete(b.cacheMap, k) // this needs a reverse lookup too.
				b.mruCache.Remove(n)
		}*/

}
