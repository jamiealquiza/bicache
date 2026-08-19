// Package bicache implements a two-tier MFU/MRU
// cache with sharded cache units.
package bicache

import (
	"context"
	"errors"
	"log"
	"math"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jamiealquiza/bicache/v2/sll"
	"github.com/jamiealquiza/tachymeter"
)

// nearestExpireSentinel is a far-future offset used as the
// nearest expire when no TTL'd keys are present; it defers
// TTL eviction scans until a SetTTL sets a real timestamp.
const nearestExpireSentinel = time.Duration(math.MaxInt32) * time.Second

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

// counters holds Bicache performance
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
	Context    context.Context
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
	MFUSize    uint   // Number of active MFU keys.
	MRUSize    uint   // Number of active MRU keys.
	MFUUsedP   uint   // MFU used in percent.
	MRUUsedP   uint   // MRU used in percent.
	MFUMaxSize uint   // Maximum number of MFU keys.
	MRUMaxSize uint   // Maximum number of MRU keys.
	Hits       uint64 // Cache hits.
	Misses     uint64 // Cache misses.
	Evictions  uint64 // Cache evictions.
	Overflows  uint64 // Failed sets on full caches.
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
	shardCount := c.ShardCount
	if shardCount == 0 {
		shardCount = 512
	}

	shards := make([]*Shard, shardCount)

	// Get cache sizes for each shard.
	mfuSize := int(math.Ceil(float64(c.MFUSize) / float64(shardCount)))
	mruSize := int(math.Ceil(float64(c.MRUSize) / float64(shardCount)))

	// Init shards.
	for i := 0; i < shardCount; i++ {
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

	parent := c.Context
	if parent == nil {
		parent = context.Background()
	}
	ctx, cf := context.WithCancel(parent)

	cache := &Bicache{
		shards:     shards,
		ShardCount: uint32(shardCount),
		Size:       (mfuSize + mruSize) * shardCount,
		done:       cf,
	}

	// Initialize a background goroutine
	// for handling promotions and evictions,
	// if configured.
	if c.AutoEvict > 0 {
		cache.autoEvict = true
		iter := time.Duration(c.AutoEvict) * time.Millisecond
		go bgAutoEvict(ctx, cache, iter, c.EvictLog)
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
func bgAutoEvict(ctx context.Context, b *Bicache, iter time.Duration, evictLog bool) {
	ttlTachy := tachymeter.New(&tachymeter.Config{Size: int(b.ShardCount)})
	promoTachy := tachymeter.New(&tachymeter.Config{Size: int(b.ShardCount)})
	interval := time.NewTicker(iter)
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
				if evictLog {
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
				s.RLock()
				nearestExpire := s.nearestExpire
				s.RUnlock()

				if nearestExpire.Before(start.Add(iter)) {
					evicted = s.evictTTL()
				}

				if evictLog && evicted > 0 {
					ttlTachy.AddTime(time.Since(start))
				}

				// Run promotions/overflow evictions.
				start = time.Now()
				s.promoteEvict()

				if evictLog {
					promoTachy.AddTime(time.Since(start))
				}
			}

			// Calc eviction/promo stats.
			ttlStats = ttlTachy.Calc()
			promoStats = promoTachy.Calc()

			if evictLog {
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

	stats.MFUMaxSize = uint(mfuCap)
	stats.MRUMaxSize = uint(mruCap)

	stats.MRUUsedP = uint(float64(stats.MRUSize) / mruCap * 100)
	// Prevent incorrect stats in MRU-only mode.
	if mfuCap > 0 {
		stats.MFUUsedP = uint(float64(stats.MFUSize) / mfuCap * 100)
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

	// Marked expirations.
	var expired []string

	// Set initial nearest expire.
	nearestExpire := time.Now().Add(nearestExpireSentinel)

	s.RLock()

	now := time.Now()
	for k, ttl := range s.ttlMap {
		if now.After(ttl) {
			expired = append(expired, k)
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
	now = time.Now()
	for _, key := range expired {
		// Recheck the TTL under the write lock; a
		// concurrent SetTTL may have extended it since
		// the mark phase, or a Del may have removed the key.
		ttl, hasTTL := s.ttlMap[key]
		if !hasTTL {
			continue
		}

		if now.Before(ttl) {
			// The key was extended; it's now eligible
			// for the nearest expire value instead.
			if ttl.Before(nearestExpire) {
				nearestExpire = ttl
			}
			continue
		}

		delete(s.ttlMap, key)
		evicted++

		if n, exists := s.cacheMap[key]; exists {
			delete(s.cacheMap, key)
			switch n.state {
			case 0:
				s.mruCache.Remove(n.node)
			case 1:
				s.mfuCache.Remove(n.node)
			}
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

	// Update the TTL and eviction counters.
	s.decrementTTLCount(uint64(evicted))
	atomic.AddUint64(&s.counters.evictions, uint64(evicted))

	return evicted
}

// promoteEvict checks if the MRU exceeds the
// Config.MRUSize (overflow count) If so, the top <overflow count>
// MRU scores are checked against the MFU. If any of the top MRU scores
// are greater than the lowest MFU scores, they are promoted
// to the MFU (if possible). Any remaining overflow count
// is evicted from the tail of the MRU.
func (s *Shard) promoteEvict() {
	// The shard is locked for the full promotion/eviction
	// pass; HighScores/LowScores traverse the cache lists
	// and the nodes they return must not be concurrently
	// removed (e.g. by a Del or TTL eviction).
	s.Lock()
	defer s.Unlock()

	// How far over MRU capacity are we?
	mruOverflow := int(s.mruCache.Len()) - int(s.mruCap)
	if mruOverflow <= 0 {
		return
	}

	// If MFU cap is 0, shortcut to
	// LRU-only behavior.
	if s.mfuCap == 0 {
		s.evictFromMRUTail(mruOverflow)
		return
	}

	// Get the top n MRU elements
	// where n = MRU capacity overflow,
	// in descending score order.
	candidates := s.mruCache.HighScores(mruOverflow)
	sort.Sort(sort.Reverse(candidates))

	// Promote as many candidates as possible
	// into free MFU slots.
	promoted := s.promoteToFreeSlots(candidates)
	mruOverflow -= promoted
	if mruOverflow == 0 {
		return
	}

	// The MFU is full; promote any remaining candidates
	// that outscore the lowest-scored MFU keys, demoting
	// those into the MRU.
	s.promoteByScore(candidates[promoted:])

	// Evict the remaining overflow from the MRU tail.
	// Score-based promotions demote the replaced MFU node
	// back into the MRU, so they don't reduce the overflow.
	s.evictFromMRUTail(mruOverflow)
}

// promoteToFreeSlots promotes candidate MRU nodes into
// unused MFU slots, in order, skipping the whole batch at
// the first low-scored candidate (candidates must be in
// descending score order). Free slots are only likely to
// exist in a somewhat new cache. The number of promotions
// is returned. The shard lock must be held.
func (s *Shard) promoteToFreeSlots(candidates sll.NodeScoreList) int {
	mfuFree := int(s.mfuCap) - int(s.mfuCache.Len())
	if mfuFree > len(candidates) {
		mfuFree = len(candidates)
	}

	var promoted int
	for i := 0; i < mfuFree; i++ {
		// Don't promote keys with low scores.
		if candidates[i].Score < 2 {
			break
		}

		s.promoteToMFU(candidates[i])
		promoted++
	}

	return promoted
}

// promoteByScore promotes candidate MRU nodes that outscore
// the lowest-scored MFU nodes, demoting each replaced MFU
// node to the MRU head. Candidates must be in descending
// score order. The shard lock must be held.
func (s *Shard) promoteByScore(candidates sll.NodeScoreList) {
	// Bottom MFU scores in ascending order.
	bottomMFU := s.mfuCache.LowScores(len(candidates))

	// Compare the highest-scored remaining candidate against
	// the lowest remaining MFU score; both lists are sorted,
	// so the first ineligible pair ends the scan.
	for i, mruNode := range candidates {
		if i == len(bottomMFU) || mruNode.Score <= bottomMFU[i].Score {
			break
		}

		s.demoteToMRU(bottomMFU[i])
		s.promoteToMFU(mruNode)
	}
}

// promoteToMFU moves an MRU-resident node to the
// MFU tail. The shard lock must be held.
func (s *Shard) promoteToMFU(n *sll.Node) {
	s.mruCache.Remove(n)
	s.mfuCache.PushTailNode(n)
	s.cacheMap[n.Value.(*cacheData).k].state = 1
}

// demoteToMRU moves an MFU-resident node to the
// MRU head. The shard lock must be held.
func (s *Shard) demoteToMRU(n *sll.Node) {
	s.mfuCache.Remove(n)
	s.mruCache.PushHeadNode(n)
	s.cacheMap[n.Value.(*cacheData).k].state = 0
}

// evictFromMRUTail evicts n keys from the tail
// of the MRU cache. The shard lock must be held.
func (s *Shard) evictFromMRUTail(n int) {
	ttlStart := len(s.ttlMap)

	var evicted int
	for i := 0; i < n; i++ {
		node := s.mruCache.Tail()
		if node == nil {
			break
		}

		k := node.Value.(*cacheData).k
		delete(s.cacheMap, k)
		delete(s.ttlMap, k)
		s.mruCache.Remove(node)
		evicted++
	}

	// Update the TTL and eviction counters.
	ttlEvicted := ttlStart - len(s.ttlMap)
	s.decrementTTLCount(uint64(ttlEvicted))
	atomic.AddUint64(&s.counters.evictions, uint64(evicted))
}

// decrementTTLCount decrements the Shard.ttlCount
// value by n. Even though these operations are atomic,
// this method should only be called when the shard is locked
// for other consistency reasons.
func (s *Shard) decrementTTLCount(n uint64) {
	// Prevents some obscure
	// scenario where ttlCount is
	// already 0 and we rollover to
	// uint max.
	if n > atomic.LoadUint64(&s.ttlCount) {
		atomic.StoreUint64(&s.ttlCount, 0)
	} else {
		atomic.AddUint64(&s.ttlCount, ^(n - 1))
	}
}
