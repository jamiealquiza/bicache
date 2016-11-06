package bicache

import (
	"sync/mutex"

	"github.com/jamiealquiza/bicache/sll"
)

// Bicache
type Bicache struct {
	sync.Mutex
	mfuCache map[string]*Sll.Node
	mruCache map[string]*Sll.Node
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