// Package bicache implements a two-tier MFU/MRU
// cache with sharded cache units.
package bicache

import (
	"container/list"
	"context"
	"errors"
	"log"
	"math"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jamiealquiza/bicache/sll"
	"github.com/jamiealquiza/tachymeter"
)

// Bicache implements a two-tier MFU/MRU
// cache with sharded cache units.
type Bicache struct {
	shards     []*Shard
	autoEvict  bool
	ShardCount uint32
	Size       int
	paused     uint32
	done       context.CancelFunc
}

// Shard implements a cache unit
// with isolated MFU/MRU caches.
type Shard struct {
	sync.RWMutex
	cacheMap      map[string]*entry
	mfuCache      *sll.Sll
	mruCache      *sll.Sll
	mfuCap        uint
	mruCap        uint
	autoEvict     bool
	ttlCount      uint64
	ttlMap        map[string]time.Time
	counters      *counters
	nearestExpire time.Time
	noOverflow    bool
}

// Counters holds Bicache performance
// data.
type counters struct {
	hits      uint64
	misses    uint64
	evictions uint64
	overflows uint64
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
	MFUSize    uint
	MRUSize    uint
	AutoEvict  uint
	EvictLog   bool
	ShardCount int
	NoOverflow bool
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
	k string
	v interface{}
}

// Stats holds Bicache
// statistics data.
type Stats struct {
	MFUSize   uint   // Number of acive MFU keys.
	MRUSize   uint   // Number of active MRU keys.
	MFUUsedP  uint   // MFU used in percent.
	MRUUsedP  uint   // MRU used in percent.
	Hits      uint64 // Cache hits.
	Misses    uint64 // Cache misses.
	Evictions uint64 // Cache evictions.
	Overflows uint64 // Failed sets on full caches.
}

// New takes a *Config and returns
// an initialized *Bicache.
func New(c *Config) (*Bicache, error) {
	// Check that ShardCount is a power of 2.
	if (c.ShardCount & (c.ShardCount - 1)) != 0 {
		return nil, errors.New("Shard count must be a power of 2")
	}

	if c.MRUSize <= 0 {
		return nil, errors.New("MRU size must be > 0")
	}

	// Default to 512 if unset.
	if c.ShardCount == 0 {
		c.ShardCount = 512
	}

	shards := make([]*Shard, c.ShardCount)

	// Get cache sizes for each shard.
	mfuSize := int(math.Ceil(float64(c.MFUSize) / float64(c.ShardCount)))
	mruSize := int(math.Ceil(float64(c.MRUSize) / float64(c.ShardCount)))

	// Init shards.
	for i := 0; i < c.ShardCount; i++ {
		shards[i] = &Shard{
			cacheMap:      make(map[string]*entry, mfuSize+mruSize),
			mfuCache:      sll.New(),
			mruCache:      sll.New(),
			mfuCap:        uint(mfuSize),
			mruCap:        uint(mruSize),
			ttlMap:        make(map[string]time.Time),
			counters:      &counters{},
			nearestExpire: time.Now(),
			noOverflow:    c.NoOverflow,
		}
	}

	ctx, cf := context.WithCancel(context.Background())

	cache := &Bicache{
		shards:     shards,
		ShardCount: uint32(c.ShardCount),
		Size:       (mfuSize + mruSize) * c.ShardCount,
		done:       cf,
	}

	// Initialize a background goroutine
	// for handling promotions and evictions,
	// if configured.
	if c.AutoEvict > 0 {
		cache.autoEvict = true
		iter := time.Duration(c.AutoEvict)
		go bgAutoEvict(ctx, cache, iter, c)
	}

	return cache, nil
}

// Close stops background tasks and
// releases any resources. This should be
// called before removing a reference to
// a *Bicache if it's desired to be garbage
// collected cleanly.
func (b *Bicache) Close() {
	b.done()
}

// bgAutoEvict calls evictTTL and promoteEvict for all shards
// sequentially on the configured iter time interval.
func bgAutoEvict(ctx context.Context, b *Bicache, iter time.Duration, c *Config) {
	ttlTachy := tachymeter.New(&tachymeter.Config{Size: c.ShardCount})
	promoTachy := tachymeter.New(&tachymeter.Config{Size: c.ShardCount})
	interval := time.NewTicker(time.Millisecond * iter)
	var evicted int
	var start time.Time

	defer interval.Stop()

	var ttlStats, promoStats *tachymeter.Metrics

	for {
		select {
		case <-ctx.Done():
			return
		case <-interval.C:
			// Skip this interval if
			// evictions are paused.
			if atomic.LoadUint32(&b.paused) == 1 {
				if c.EvictLog {
					log.Printf("[Bicache] Evictions Paused")
				}
				continue
			}

			// On the auto eviction interval,
			// we loop through each shard
			// and trigger a TTL and promotion/eviction.
			for _, s := range b.shards {
				// Run ttl evictions.
				start = time.Now()
				evicted = 0

				// At the very first check, nearestExpire
				// was set to the Bicache initialization time.
				// This is certain to run at least once.
				// The first and real nearest expire will be set
				// in any SetTTL call that's made.
				if s.nearestExpire.Before(start.Add(iter)) {
					evicted = s.evictTTL()
				}

				if c.EvictLog && evicted > 0 {
					ttlTachy.AddTime(time.Since(start))
				}

				// Run promotions/overflow evictions.
				start = time.Now()
				s.promoteEvict()

				if c.EvictLog {
					promoTachy.AddTime(time.Since(start))
				}
			}

			// Calc eviction/promo stats.
			ttlStats = ttlTachy.Calc()
			promoStats = promoTachy.Calc()

			if c.EvictLog {
				// Log TTL stats if a
				// TTL eviction was triggered.
				if ttlStats.Count > 0 {
					log.Printf("[Bicache EvictTTL] cumulative: %s | min: %s | max: %s\n",
						ttlStats.Time.Cumulative, ttlStats.Time.Min, ttlStats.Time.Max)
				}

				// Log PromoteEvict stats.
				log.Printf("[Bicache PromoteEvict] cumulative: %s | min: %s | max: %s\n",
					promoStats.Time.Cumulative, promoStats.Time.Min, promoStats.Time.Max)
			}

			// Reset tachymeter.
			ttlTachy.Reset()
			promoTachy.Reset()
		}
	}
}

// Stats returns a *Stats with
// Bicache statistics data.
func (b *Bicache) Stats() *Stats {
	stats := &Stats{}
	var mfuCap, mruCap float64

	for _, s := range b.shards {
		s.RLock()
		stats.MFUSize += s.mfuCache.Len()
		stats.MRUSize += s.mruCache.Len()
		s.RUnlock()

		mfuCap += float64(s.mfuCap)
		mruCap += float64(s.mruCap)

		stats.Hits += atomic.LoadUint64(&s.counters.hits)
		stats.Misses += atomic.LoadUint64(&s.counters.misses)
		stats.Evictions += atomic.LoadUint64(&s.counters.evictions)
		stats.Overflows += atomic.LoadUint64(&s.counters.overflows)
	}

	stats.MRUUsedP = uint(float64(stats.MRUSize) / mruCap * 100)
	// Prevent incorrect stats in MRU-only mode.
	if mfuCap > 0 {
		stats.MFUUsedP = uint(float64(stats.MFUSize) / mfuCap * 100)
	} else {
		stats.MFUUsedP = 0
	}

	return stats
}

// evictTTL evicts expired keys using a mark
// sweep garbage collection. The number of keys
// evicted is returned.
func (s *Shard) evictTTL() int {
	// Return if we have no TTL'd keys.
	if atomic.LoadUint64(&s.ttlCount) == 0 {
		return 0
	}

	// Tracking marked expirations
	// in a list.
	expired := list.New()

	// Set initial nearest expire.
	nearestExpire := time.Now().Add(time.Second * 2147483647)

	s.RLock()

	now := time.Now()
	for k, ttl := range s.ttlMap {
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

	s.RUnlock()

	// Lock and evict.
	s.Lock()

	var evicted int
	for k := expired.Front(); k != nil; k = k.Next() {
		if n, exists := s.cacheMap[k.Value.(string)]; exists {
			delete(s.cacheMap, k.Value.(string))
			delete(s.ttlMap, k.Value.(string))
			switch n.state {
			case 0:
				s.mruCache.Remove(n.node)
			case 1:
				s.mfuCache.Remove(n.node)
			}
			evicted++
		}
	}

	// Update the nearest expire.
	// If the last TTL'd key was just expired,
	// this will be left at the initially set value
	// at the top of evictTTL. This means that the
	// auto eviction runs will just skip
	// evictTTL until a SetTTL creates a real
	// nearest expire timestamp (since it's checking
	// if the nearest expire happens within the auto
	// evict interval).
	s.nearestExpire = nearestExpire

	s.Unlock()

	// Update eviction counters.
	s.decrementTTLCount(uint64(evicted))

	return evicted
}

// promoteEvict checks if the MRU exceeds the
// Config.MRUSize (overflow count) If so, the top <overflow count>
// MRU scores are checked against the MFU. If any of the top MRU scores
// are greater than the lowest MFU scores, they are promoted
// to the MFU (if possible). Any remaining overflow count
// is evicted from the tail of the MRU.
func (s *Shard) promoteEvict() {
	// How far over MRU capacity are we?
	mruOverflow := int(s.mruCache.Len() - s.mruCap)
	if mruOverflow <= 0 {
		return
	}

	// If MFU cap is 0, shortcut to
	// LRU-only behavior.
	if s.mfuCap == 0 {
		s.Lock()
		s.evictFromMRUTail(mruOverflow)
		s.Unlock()

		return
	}

	// Get the top n MRU elements
	// where n = MRU capacity overflow.
	mruToPromoteEvict := s.mruCache.HighScores(mruOverflow)

	// HighScores is expensive.
	// Get lock immediately after.
	// May want to sanity check that
	// keys weren't removed during.
	s.Lock()

	// Reverse into descending order.
	sort.Sort(sort.Reverse(mruToPromoteEvict))

	// Check MFU capacity.
	mfuFree := int(s.mfuCap - s.mfuCache.Len())
	if mfuFree < 0 {
		mfuFree = 0
	}

	var promoted int

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
			// Don't promote keys with low scores.
			// We can break since the mruToPromoteEvict
			// list is in descending order.
			if node.Score < 2 {
				break
			}
			// Remove from the MRU and
			// push to the MFU tail.
			// Update cache state.
			s.mruCache.Remove(node)
			s.mfuCache.PushTailNode(node)
			s.cacheMap[node.Value.(*cacheData).k].state = 1

			promoted++
		}

		// If we were able to promote
		// all the overflow, return.
		if promoted == mruOverflow {
			s.Unlock()
			return
		}
	}

promoteByScore:
	s.Unlock()
	// Get a remainder to either promote by score
	// to the MFU or ultimately evict from the MRU.
	mruOverflow -= promoted
	remainderPosition := promoted

	// Some vars are declared up here
	// due to the goto jumps.

	// Counter to track
	// how many from the MRU keys
	// were promoted by score.
	var promotedByScore int

	// We're here on two conditions:
	// 1) The MFU was full. We need to handle all mruToPromoteEvict (canPromote == 0).
	// 2) We promoted some mruToPromoteEvict and have leftovers (canPromote > 0).

	// Get top MRU scores and bottom MFU scores to compare.
	bottomMFU := s.mfuCache.LowScores(mruOverflow)

	// If the lowest MFU score is higher than the lowest
	// score to promote, none of these are eligible.
	if len(bottomMFU) == 0 || bottomMFU[0].Score >= mruToPromoteEvict[remainderPosition].Score {
		goto evictFromMRUTail
	}

	// Otherwise, scan for a replacement.
	s.Lock()
scorePromote:
	for _, mruNode := range mruToPromoteEvict[remainderPosition:] {
		for i, mfuNode := range bottomMFU {
			if mruNode.Score > mfuNode.Score {
				// Push the evicted MFU node to the head
				// of the MRU and update state.
				s.mfuCache.Remove(mfuNode)
				s.mruCache.PushHeadNode(mfuNode)
				s.cacheMap[mfuNode.Value.(*cacheData).k].state = 0

				// Promote the MRU node to the MFU and
				// update state.
				s.mruCache.Remove(mruNode)
				s.mfuCache.PushTailNode(mruNode)
				s.cacheMap[mruNode.Value.(*cacheData).k].state = 1

				promotedByScore++

				// Remove the replaced MFU node from the
				// bottomMFU list so it's not attempted twice.
				bottomMFU = append(bottomMFU[:i], bottomMFU[i+1:]...)
				break
			}
			if i == len(bottomMFU)-1 {
				break scorePromote
			}
		}

	}

	s.Unlock()

evictFromMRUTail:

	s.Lock()

	// What's the overflow remainder count?
	toEvict := mruOverflow - promotedByScore
	// Evict this many from the MRU tail.
	if toEvict > 0 {
		s.evictFromMRUTail(toEvict)
	}

	s.Unlock()
}

// evictFromMRUTail evicts n keys from the tail
// of the MRU cache.
func (s *Shard) evictFromMRUTail(n int) {
	ttlStart := len(s.ttlMap)

	for i := 0; i < n; i++ {
		node := s.mruCache.Tail()
		delete(s.cacheMap, node.Value.(*cacheData).k)
		delete(s.ttlMap, node.Value.(*cacheData).k)
		s.mruCache.RemoveTail()
	}

	// Update the ttlCount.
	ttlEvicted := ttlStart - len(s.ttlMap)
	s.decrementTTLCount(uint64(ttlEvicted))
	// Update eviction count.
	// Excludes TTL evictions since the
	// decrementTTLCount handles that for us.
	atomic.AddUint64(&s.counters.evictions, uint64(n-ttlEvicted))
}

// decrementTTLCount decrements the Bicache.ttlCount
// value by n. Even though these operations are atomic,
// this method should only be called when the shard is locked
// for other consistency reasons.
func (s *Shard) decrementTTLCount(n uint64) {
	// Prevents some obscure
	// scenario where ttlCount is
	// already 0 and we rollover to
	// uint max.
	if s.ttlCount-n > s.ttlCount {
		atomic.StoreUint64(&s.ttlCount, 0)
	} else {
		atomic.AddUint64(&s.ttlCount, ^uint64(n-1))
	}

	// Increment the evictions count
	// by n, regardless.
	atomic.AddUint64(&s.counters.evictions, n)
}
