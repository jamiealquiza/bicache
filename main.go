package bicache

import (
	"sync/mutex"

	"github.com/jamiealquiza/bicache/sll"
)

// Bicache
type Bicache struct {
	sync.Mutex
	mfuCache map[Object]string // Val may be a ts for ttl/gc.
	mruCache map[Object]string
	mfuCap int
	mruCap int
	safe bool
}

// Config
type Config struct {
	MruSize int
	MfuSize int
	Safe bool
	// Inverted index
}

// New
func New(c *Config) *Bicache {
	return &Bicache {
		mfuCache: make(map[string]string),
		mruCache: make(map[string]string),
		mfuCap: c.MfuSize,
		mruCap: c.MruSize,
		ll: list.New(),
	}
}

// Set
func (b *Bicache) Set(k, v string) {
	if b.safe {
		b.Lock()
		defer b.Unlock()
	}


}

// Get
func (b *Bicache) Get(k string) string {
	if b.safe {
		b.Lock()
		defer b.Unlock()
	}


}