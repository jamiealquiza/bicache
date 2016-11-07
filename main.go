package bicache

import (
	"sync"

	"github.com/jamiealquiza/bicache/sll"
)

// Bicache
type Bicache struct {
	sync.RWMutex
	mfuCacheMap map[interface{}]*sll.Node
	mruCacheMap map[interface{}]*sll.Node
	mfuCache    *sll.Sll
	mruCache    *sll.Sll
	mfuCap      int
	mruCap      int
}

// Config
type Config struct {
	MruSize int
	MfuSize int
	// Inverted index
}

// New
func New(c *Config) *Bicache {
	return &Bicache{
		mfuCacheMap: make(map[interface{}]*sll.Node),
		mruCacheMap: make(map[interface{}]*sll.Node),
		mfuCache:    sll.New(),
		mruCache:    sll.New(),
		mfuCap:      c.MfuSize,
		mruCap:      c.MruSize,
	}
}

// Set
func (b *Bicache) Set(k, v interface{}) {
	b.Lock()
	defer b.Unlock()

	// Check the MFU cache. We never manually
	// create an entry in the MFU; only update
	// an existing entry.
	if n, exists := b.mfuCacheMap[k]; exists {
		n.Value = v
		return
	}

	// Check the MRU cache. If it exists, update.
	// If it doesn't exist, create at head.
	if n, exists := b.mruCacheMap[k]; exists {
		n.Value = v
		b.mruCache.MoveToHead(n)
	} else {
		n = b.mruCache.PushHead(v)
		b.mruCacheMap[k] = n
	}

	// Check MRU size here.
}

// Get
func (b *Bicache) Get(k interface{}) interface{} {
	b.RLock()
	defer b.RUnlock()

	if n, exists := b.mruCacheMap[k]; !exists {
		return nil
	} else {
		return n.Read()
	}
}

// Delete
//func (b *Bicache) Delete(k interface{}) {