package bicache

import (
	//"fmt"
	"sync"

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
	cacheMap map[interface{}]*entry
	mfuCache *sll.Sll
	mruCache *sll.Sll
	mfuCap   uint
	mruCap   uint
	// MFU top/bottom scores.
}

// Config holds a Bicache configuration.
// The MFU and MRU cache sizes are set in number
// of keys.
type Config struct {
	MfuSize uint
	MruSize uint
	// DeferEviction true // on-write vs automatic
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
	return &Bicache{
		cacheMap: make(map[interface{}]*entry),
		mfuCache: sll.New(),
		mruCache: sll.New(),
		mfuCap:   c.MfuSize,
		mruCap:   c.MruSize,
	}
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

// Set takes a key and value and creates
// and entry in the MRU cache. If the key
// already exists, the value is updated.
func (b *Bicache) Set(k, v interface{}) {
	b.Lock()
	// If the entry exists, update. If not,
	// create at the tail of the MRU cache.
	if n, exists := b.cacheMap[k]; exists {
		n.node.Value = v
		if n.state == 0 {
			b.mruCache.MoveToHead(n.node)
		}
	} else {
		// Create at the MRU tail.
		newNode := b.mruCache.PushTail(v)
		b.cacheMap[k] = &entry{
			node:  newNode,
			state: 0,
		}
	}

	b.Unlock()
	b.PromoteEvict()
}

// Get takes a key and returns the value. Every get
// on a key increases the key score.
func (b *Bicache) Get(k interface{}) interface{} {
	b.RLock()
	defer b.RUnlock()

	if n, exists := b.cacheMap[k]; exists {
		return n.node.Read()
	}

	return nil
}

// Delete deletes a key.
func (b *Bicache) Delete(k interface{}) {
	b.Lock()
	defer b.Unlock()

	if n, exists := b.cacheMap[k]; exists {
		delete(b.cacheMap, k)
		switch n.state {
		case 0:
			b.mruCache.Remove(n.node)
		case 1:
			b.mfuCache.Remove(n.node)
		}
	}
}

// PromoteEvict checks if the MRU exceeds the
// Config.MruSize. If so, the top MRU scores are
// checked against the MFU. If any of the top MRU scores
// are greater than the lowest MFU scores, they are promoted
// to the MFU. Any remaining count of evictions that must occur
// are removed from the tail of the MRU.
func (b *Bicache) PromoteEvict() {
	b.Lock()
	defer b.Unlock()

	// How far over capacity are we?
	mruOverflow := int(b.mruCache.Len() - b.mruCap)
	if mruOverflow > 0 {
		topMru := b.mruCache.HighScores(mruOverflow)
		// Move overflow to MFU.
		for _, node := range topMru {
			//b.cacheMap[].state == 1
			// Need to find the node in the cacheMap and
			// update the state.
			b.mruCache.Remove(node)
			b.mfuCache.PushTailNode(node)
		}

	}
}
