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
func (n *Node) Read() interface{} {
	n.Score++
	return n.Value
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
func (ll *Sll) MoveToHead(n *Node) {
	ll.Lock()
	defer ll.Unlock()

	// If this is the tail, update
	// assign a new tail.
	if n.Prev == nil {
		ll.tail = n.Next
	}

	// Set current head next to n.
	ll.head.Next = n
	// Set n prev to current head.
	n.Prev = ll.head
	// Swap n to head.
	ll.head = n
	// Ensure head.Next is nil.
	ll.head.Next = nil
}

// MoveToTail takes a *Node and moves it
// to the back of the *Sll.
func (ll *Sll) MoveToTail(n *Node) {
	ll.Lock()
	defer ll.Unlock()

	// If this is the head, update
	// assign a new head.
	if n.Next == nil {
		ll.head = n.Prev
	}

	// Set current tail prev to n.
	ll.tail.Prev = n
	// Set n next to current tail.
	n.Next = ll.tail
	// Swap n to tail.
	ll.tail = n
	// Ensure tail.Prev is nil.
	ll.tail.Prev = nil
}

// PushHead creates a *Node with value v
// at the head of the *Sll and returns a *Node.
func (ll *Sll) PushHead(v interface{}) *Node {
	ll.Lock()
	defer ll.Unlock()

	n := &Node{
		Value: v,
		Score: 0,
	}

	ll.scores = append(ll.scores, n)

	// Is this a new ll?
	if ll.head == nil {
		ll.head = n
		ll.tail = n
		return n
	}

	// Set current head next to n.
	ll.head.Next = n
	// Set n prev to current head.
	n.Prev = ll.head
	// Swap n to head.
	ll.head = n
	// Ensure head.Next is nil.
	ll.head.Next = nil

	return n
}

// PushTail creates a *Node with value v
// at the tail of the *Sll and returns a *Node.
func (ll *Sll) PushTail(v interface{}) *Node {
	ll.Lock()
	defer ll.Unlock()

	n := &Node{
		Value: v,
		Score: 0,
	}

	ll.scores = append(ll.scores, n)

	// Is this a new ll?
	if ll.tail == nil {
		ll.head = n
		ll.tail = n
		return n
	}

	// Set current tail prev to n.
	ll.tail.Prev = n
	// Set n next to current tail.
	n.Next = ll.tail
	// Swap n to tail.
	ll.tail = n
	// Ensure tail.Prev is nil.
	ll.tail.Prev = nil

	return n
}

// Remove removes a *Node from the *Sll.
func (ll *Sll) Remove(n *Node) {
	ll.Lock()
	defer ll.Unlock()

	// Link next and prev.
	n.Next.Prev = n.Prev
	n.Prev.Next = n.Next

	// Remove references.
	n.Next, n.Prev = nil, nil
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
