package bicache

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

	if !b.autoEvict {
		b.PromoteEvict()
	}
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
