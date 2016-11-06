package bicache

import (
	"sync"

	"github.com/jamiealquiza/bicache/sll"
)

// Bicache
type Bicache struct {
	sync.Mutex
	mfuCacheMap map[interface{}]*sll.Node
	mruCacheMap map[interface{}]*sll.Node
	mfuCache    *sll.Sll
	mruCache    *sll.Sll
	mfuCap      int
	mruCap      int
	safe        bool
}

// Config
type Config struct {
	MruSize int
	MfuSize int
	Safe    bool
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
		safe:        c.Safe,
	}
}

// Set
func (b *Bicache) Set(k, v interface{}) {
	if b.safe {
		b.Lock()
		defer b.Unlock()
	}

	if n, exists := b.mruCacheMap[k]; !exists {
		n = b.mruCache.PushHead(v)
		b.mruCacheMap[k] = n
	} else {
		n.Value = v
		b.mruCache.MoveToHead(n)
	}
}

// Get
func (b *Bicache) Get(k interface{}) interface{} {
	if b.safe {
		b.Lock()
		defer b.Unlock()
	}

	if n, exists := b.mruCacheMap[k]; !exists {
		return nil
	} else {
		return n.Read()
	}
}
