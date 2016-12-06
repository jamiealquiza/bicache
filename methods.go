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

// Get takes a key and returns the value. Every get
// on a key increases the key score.
func (b *Bicache) Get(k interface{}) interface{} {
	b.RLock()
	defer b.RUnlock()

	if n, exists := b.cacheMap[k]; exists {
		read := n.node.Read()
		return read.(*cacheData).v
	}

	return nil
}

// Del deletes a key.
func (b *Bicache) Del(k interface{}) {
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
