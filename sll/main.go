package sll

import (
	"sync/mutex"
)

type Sll struct {
	sync.Mutex
	Head *Object
	Tail *Object
	Len int
	Scores ObjectScoreList
}

// Object
type Object struct {
	Value interface{}
	HitCount int64
	Next *Object
	Prev *Object
	// Might add a create for TTL.
}

// New
func New() *Sll {
	return &Sll{}
}

// Satisfy sort interface.

type ObjectScoreList []*Object

func (osl ObjectScoreList) Len() int {
	return len(osl)
}

func (osl ObjectScoreList) Less(i, j int) bool {
	return osl[i].HitCount < osl[j].HitCount
}

func (osl ObjectScoreList) Swap(i, j int) {
	osl[i], osl[j] = osl[j], osl[i]
}

func (ll *Sll) HighScore() *Object {

}

func (ll *Sll) LowScore() *Object {
	
}

// PushHead
func (ll *Sll) PushHead(v interface{}) {
	ll.Lock()
	defer ll.Unlock()

	o := &Object{
		Value: v,
		HitCount: 1,
	}

	// If no head object.
	if ll.Head == nil {
		ll.Head = o
		return
	}

	// Set current head next to o.
	ll.Head.Next = o
	// Set o prev to current head.
	o.Prev = ll.Head
	// Swap o to head.
	ll.Head = o
	// Ensure head.Next is nil.
	ll.Head.Next = nil

	// insort here.

	ll.Len = len(ll.Scores)
}

// PushTail
func (ll *Sll) PushTail(v interface{}) {
	ll.Lock()
	defer ll.Unlock()

	o := &Object{
		Value: v,
		HitCount: 1,
	}

	// If no tail object.
	if ll.Tail == nil {
		ll.Tail = o
		return
	}

	// Set current tail prev to o.
	ll.Tail.Prev = o
	// Set o next to current tail.
	o.Next = ll.Tail
	// Swap o to tail.
	ll.Tail = o
	// Ensure tail.Prev is nil.
	ll.Tail.Prev = nil
}

// Remove
func (ll *Sll) Remove(o *Object) {
	ll.Lock()
	defer ll.Unlock()

	// Link next and prev.
	o.Next.Prev = o.Prev
	o.Prev.Next = o.Next

	// Remove references.
	o.Next, o.Prev = nil
}

// RemoveHead
func (ll *Sll) RemoveHead() {
	ll.Lock()
	defer ll.Unlock()

	if ll.Head == nil {
		return
	}

	// Set head to current head.Prev.
	ll.Head = ll.Head.Prev
	// Unlink old head's refs.
	ll.Head.Next.Prev = nil
	// Set new head.Next to nil.
	ll.Head.Next = nil
}

// RemoveTail
func (ll *Sll) RemoveTail() {
	ll.Lock()
	defer ll.Unlock()

	if ll.Tail == nil {
		return
	}

	// Set tail to current tail.Next
	ll.Tail = ll.Tail.Next
	// Unlink old tail's refs.
	ll.Tail.Prev.Next = nil
	// Set new tail.Prev to nil.
	ll.Tail.Prev = nil
}