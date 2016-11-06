package sll

import (
	"sort"
	"sync"
)

// Sll is a scored linked list.
type Sll struct {
	sync.Mutex
	head   *Node
	tail   *Node
	scores nodeScoreList
}

// Node is a scored linked list node.
type Node struct {
	Value interface{}
	Score int64
	Next  *Node
	Prev  *Node
	// Might add a create field for TTL.
}

// New creates a new *Sll
func New() *Sll {
	return &Sll{
		scores: nodeScoreList{},
	}
}

// nodeScoreList holds a slice of *Node
// for sorting by score.
type nodeScoreList []*Node

// Read returns a *Node Value and increments the score.
func (o *Node) Read() interface{} {
	o.Score++
	return o.Value
}

// nodeScoreList methods to satisfy the sort interface.

func (nsl nodeScoreList) Len() int {
	return len(nsl)
}

func (nsl nodeScoreList) Less(i, j int) bool {
	return nsl[i].Score < nsl[j].Score
}

func (nsl nodeScoreList) Swap(i, j int) {
	nsl[i], nsl[j] = nsl[j], nsl[i]
}

// Len returns the count of nodes in the *Sll.
func (ll *Sll) Len() int {
	return len(ll.scores)
}

// Head returns the head *Node.
func (ll *Sll) Head() *Node {
	return ll.head
}

// Tail returns the head *Node.
func (ll *Sll) Tail() *Node {
	return ll.tail
}

// HighScores takes an integer and returns the
// respective number of *Nodes with the higest scores
// sorted in ascending order. The last element will
// be the node with the highest score. Calling HighScores
// locks the *Sll for the duration of a binary sort
// of roughly O(log ll.Len()) time.
func (ll *Sll) HighScores(r int) []*Node {
	ll.Lock()
	defer ll.Unlock()

	sort.Sort(ll.scores)

	if r > ll.Len() {
		return ll.scores
	}

	return ll.scores[len(ll.scores)-r:]
}

// LowScores takes an integer and returns the
// respective number of *Nodes with the lowest scores
// sorted in ascending order. The first element will
// be the node with the lowest score. Calling LowScores
// locks the *Sll for the duration of a binary sort
// of roughly O(log ll.Len()) time.
func (ll *Sll) LowScores(r int) []*Node {
	ll.Lock()
	defer ll.Unlock()

	sort.Sort(ll.scores)

	if r > len(ll.scores) {
		return ll.scores
	}

	return ll.scores[:r]
}

// MoveToHead takes a *Node and moves it
// to the front of the *Sll.
func (ll *Sll) MoveToHead(o *Node) {
	ll.Lock()
	defer ll.Unlock()

	// If this is the tail, update
	// assign a new tail.
	if o.Prev == nil {
		ll.tail = o.Next
	}

	// Set current head next to o.
	ll.head.Next = o
	// Set o prev to current head.
	o.Prev = ll.head
	// Swap o to head.
	ll.head = o
	// Ensure head.Next is nil.
	ll.head.Next = nil
}

// MoveToTail takes a *Node and moves it
// to the back of the *Sll.
func (ll *Sll) MoveToTail(o *Node) {
	ll.Lock()
	defer ll.Unlock()

	// If this is the head, update
	// assign a new head.
	if o.Next == nil {
		ll.head = o.Prev
	}

	// Set current tail prev to o.
	ll.tail.Prev = o
	// Set o next to current tail.
	o.Next = ll.tail
	// Swap o to tail.
	ll.tail = o
	// Ensure tail.Prev is nil.
	ll.tail.Prev = nil
}

// PushHead creates a *Node with value v
// at the head of the *Sll.
func (ll *Sll) PushHead(v interface{}) {
	ll.Lock()
	defer ll.Unlock()

	o := &Node{
		Value: v,
		Score: 0,
	}

	ll.scores = append(ll.scores, o)

	// Is this a new ll?
	if ll.head == nil {
		ll.head = o
		ll.tail = o
		return
	}

	// Set current head next to o.
	ll.head.Next = o
	// Set o prev to current head.
	o.Prev = ll.head
	// Swap o to head.
	ll.head = o
	// Ensure head.Next is nil.
	ll.head.Next = nil
}

// PushTail creates a *Node with value v
// at the tail of the *Sll.
func (ll *Sll) PushTail(v interface{}) {
	ll.Lock()
	defer ll.Unlock()
	
	o := &Node{
		Value: v,
		Score: 0,
	}

	ll.scores = append(ll.scores, o)

	// Is this a new ll?
	if ll.tail == nil {
		ll.head = o
		ll.tail = o
		return
	}

	// Set current tail prev to o.
	ll.tail.Prev = o
	// Set o next to current tail.
	o.Next = ll.tail
	// Swap o to tail.
	ll.tail = o
	// Ensure tail.Prev is nil.
	ll.tail.Prev = nil
}

// Remove removes a *Node from the *Sll.
func (ll *Sll) Remove(o *Node) {
	ll.Lock()
	defer ll.Unlock()

	// Link next and prev.
	o.Next.Prev = o.Prev
	o.Prev.Next = o.Next

	// Remove references.
	o.Next, o.Prev = nil, nil
}

// RemoveHead removes the current *Sll.head.
func (ll *Sll) RemoveHead() {
	ll.Lock()
	defer ll.Unlock()

	if ll.head == nil {
		return
	}

	// Set head to current head.Prev.
	ll.head = ll.head.Prev
	// Unlink old head's refs.
	ll.head.Next.Prev = nil
	// Set new head.Next to nil.
	ll.head.Next = nil
}

// RemoveTail removes the current *Sll.tail.
func (ll *Sll) RemoveTail() {
	ll.Lock()
	defer ll.Unlock()

	if ll.tail == nil {
		return
	}

	// Set tail to current tail.Next
	ll.tail = ll.tail.Next
	// Unlink old tail's refs.
	ll.tail.Prev.Next = nil
	// Set new tail.Prev to nil.
	ll.tail.Prev = nil
}
