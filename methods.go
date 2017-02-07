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
	"sync/atomic"
	"time"
)

// keyInfo holds a key name, state (0: MRU, 1: MRU)
// and cache score.
type KeyInfo struct {
	Key   interface{}
	State uint8
	Score uint64
}

// listResults is a container that holds results from
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
func (b *Bicache) Set(k, v interface{}) {
	b.Lock()
	// If the entry exists, update. If not,
	// create at the tail of the MRU cache.
	if n, exists := b.cacheMap[k]; exists {
		n.node.Value.(*cacheData).v = v
		if n.state == 0 {
			b.mruCache.MoveToHead(n.node)
		}
	} else {
		// Create at the MRU tail.
		b.cacheMap[k] = &entry{
			node: b.mruCache.PushHead(&cacheData{k: k, v: v}),
		}
	}

	b.Unlock()

	// PromoteEvict on write if it's
	// not being handled automatically.
	if !b.autoEvict {
		b.PromoteEvict()
	}
}

// SetTtl is the same as set but accepts a
// parameter t to specify a TTL in seconds.
func (b *Bicache) SetTtl(k, v interface{}, t int32) {
	b.Lock()

	// Set TTL expiration
	expiration := time.Now().Add(time.Second * time.Duration(t))
	b.ttlMap[k] = expiration

	// Increment TTL counter.
	atomic.AddUint64(&b.ttlCount, 1)

	// Proceed to normal Set operation.
	// This logic is duplicated for now
	// to skip releasing / re-acquiring a mutex.

	// If the entry exists, update. If not,
	// create at the tail of the MRU cache.
	if n, exists := b.cacheMap[k]; exists {
		n.node.Value.(*cacheData).v = v
		if n.state == 0 {
			b.mruCache.MoveToHead(n.node)
		}
	} else {
		// Create at the MRU tail.
		b.cacheMap[k] = &entry{
			node: b.mruCache.PushHead(&cacheData{k: k, v: v}),
		}
	}

	// Update the nearest expire.
	if expiration.Before(b.nearestExpire) {
		b.nearestExpire = expiration
	}

	b.Unlock()

	// PromoteEvict on write if it's
	// not being handled automatically.
	if !b.autoEvict {
		b.PromoteEvict()
	}
}

// Get takes a key and returns the value. Every get
// on a key increases the key score.
func (b *Bicache) Get(k interface{}) interface{} {
	b.RLock()

	if n, exists := b.cacheMap[k]; exists {
		read := n.node.Read()

		b.RUnlock()
		atomic.AddUint64(&b.counters.hits, 1)

		return read.(*cacheData).v
	}

	b.RUnlock()
	atomic.AddUint64(&b.counters.misses, 1)

	return nil
}

// Del deletes a key.
func (b *Bicache) Del(k interface{}) {
	b.Lock()

	if n, exists := b.cacheMap[k]; exists {
		delete(b.cacheMap, k)
		delete(b.ttlMap, k)
		switch n.state {
		case 0:
			b.mruCache.Remove(n.node)
		case 1:
			b.mfuCache.Remove(n.node)
		}
	}

	b.Unlock()
}

// List returns all key names, states, and scores
// sorted in descending order by score. Returns n
// top restults.
func (b *Bicache) List(n int) ListResults {
	b.RLock()

	// Make a ListResults large enough to hold the
	// number of cache items present in both cache tiers.
	lr := make(ListResults, b.mruCache.Len()+b.mfuCache.Len())

	var i int
	for k, v := range b.cacheMap {
		lr[i] = &KeyInfo{
			Key:   k,
			State: v.state,
			Score: v.node.Score,
		}
		i++
	}

	b.RUnlock()

	sort.Sort(lr)
	// return the number
	// of items requested.
	if n < len(lr) {
		return lr[:n]
	}

	return lr
}
