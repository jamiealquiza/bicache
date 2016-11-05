package sll

import (
	"sort"
	"sync"
)

// Sll
type Sll struct {
	sync.Mutex
	Head   *Node
	Tail   *Node
	scores NodeScoreList
}

// Node
type Node struct {
	Value interface{}
	Score int64
	Next  *Node
	Prev  *Node
	// Might add a create for TTL.
}

// New
func New() *Sll {
	return &Sll{
		scores: NodeScoreList{},
	}
}

// NodeScoreList
type NodeScoreList []*Node

// Read
func (o *Node) Read() interface{} {
	o.Score++
	return o.Value
}

// Satisfy sort interface.

func (osl NodeScoreList) Len() int {
	return len(osl)
}

func (osl NodeScoreList) Less(i, j int) bool {
	return osl[i].Score < osl[j].Score
}

func (osl NodeScoreList) Swap(i, j int) {
	osl[i], osl[j] = osl[j], osl[i]
}

func (ll *Sll) Len() int {
	return len(ll.scores)
}

// HighScores
func (ll *Sll) HighScores(r int) []*Node {
	ll.Lock()
	defer ll.Unlock()

	sort.Sort(ll.scores)

	if r > ll.Len() {
		return ll.scores
	}

	return ll.scores[len(ll.scores)-r:]
}

// LowScores
func (ll *Sll) LowScores(r int) []*Node {
	ll.Lock()
	defer ll.Unlock()

	sort.Sort(ll.scores)

	if r > len(ll.scores) {
		return ll.scores
	}

	return ll.scores[:r]
}

// MoveToHead
func (ll *Sll) MoveToHead(o *Node) {
	ll.Lock()
	defer ll.Unlock()

	// If this is the tail, update
	// assign a new tail.
	if o.Prev == nil {
		ll.Tail = o.Next
	}

	// Set current head next to o.
	ll.Head.Next = o
	// Set o prev to current head.
	o.Prev = ll.Head
	// Swap o to head.
	ll.Head = o
	// Ensure head.Next is nil.
	ll.Head.Next = nil
}

// MoveToTail
func (ll *Sll) MoveToTail(o *Node) {
	ll.Lock()
	defer ll.Unlock()

	// If this is the head, update
	// assign a new head.
	if o.Next == nil {
		ll.Head = o.Prev
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

// PushHead
func (ll *Sll) PushHead(v interface{}) {
	o := &Node{
		Value: v,
		Score: 0,
	}

	ll.scores = append(ll.scores, o)

	// Is this a new ll?
	if ll.Head == nil {
		ll.Head = o
		ll.Tail = o
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
}

// PushTail
func (ll *Sll) PushTail(v interface{}) {
	o := &Node{
		Value: v,
		Score: 0,
	}

	ll.scores = append(ll.scores, o)

	// Is this a new ll?
	if ll.Tail == nil {
		ll.Head = o
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
func (ll *Sll) Remove(o *Node) {
	ll.Lock()
	defer ll.Unlock()

	// Link next and prev.
	o.Next.Prev = o.Prev
	o.Prev.Next = o.Next

	// Remove references.
	o.Next, o.Prev = nil, nil
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
