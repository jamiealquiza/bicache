package bicache

import (
	"sync"

	"github.com/jamiealquiza/bicache/sll"
)

// Bicache
type Bicache struct {
	sync.RWMutex
	CacheMap map[interface{}]*Entry
	mfuCache    *sll.Sll
	mruCache    *sll.Sll
	mfuCap      uint
	mruCap      uint
	// MFU top/bottom scores.
}

// Config
type Config struct {
	MfuSize uint
	MruSize uint
	// DeferEviction true // on-write vs automatic
}

// Entry
type Entry struct {
	node *sll.Node
	state uint8 // 0 = MRU, 1 = MFU
}

// Stats
type Stats struct {
	MfuSize uint
	MruSize uint
	MfuUsedP uint
	MruUsedP uint
}

// New
func New(c *Config) *Bicache {
	return &Bicache{
		CacheMap: make(map[interface{}]*Entry),
		mfuCache:    sll.New(),
		mruCache:    sll.New(),
		mfuCap:      c.MfuSize,
		mruCap:      c.MruSize,
	}
}

// Stats
func (b *Bicache) Stats() *Stats {
	b.RLock()
	stats := &Stats{MfuSize: b.mfuCache.Len(), MruSize: b.mruCache.Len()}
	b.RUnlock()

	stats.MfuUsedP = uint(float64(stats.MfuSize) / float64(b.mfuCap) * 100)
	stats.MruUsedP = uint(float64(stats.MruSize) / float64(b.mruCap) * 100)

	return stats
}

// Set
func (b *Bicache) Set(k, v interface{}) {
	b.Lock()
	defer b.Unlock()

	// If the entry exists, update. If not,
	// create at the tail of the MRU cache.
	if n, exists := b.CacheMap[k]; exists {
		n.node.Value = v
		if n.state == 0 {
			b.mruCache.MoveToHead(n.node)
		}
	} else {
		// Create at the MRU tail.
		newNode := b.mruCache.PushTail(v)
		b.CacheMap[k] = &Entry{
			node: newNode,
			state: 0,
		}
	}
}

// Get
func (b *Bicache) Get(k interface{}) interface{} {
	b.RLock()
	defer b.RUnlock()

	if n, exists := b.CacheMap[k]; exists {
		return n.node.Read()
	}

	return nil
}

// Delete
func (b *Bicache) Delete(k interface{}) {
	b.Lock()
	defer b.Unlock()

	if n, exists := b.CacheMap[k]; exists {
		delete(b.CacheMap, k)
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
//func (b *Bicache) PromoteEvict() {}