package sll

import (
	"container/heap"
	"sort"
	"sync/atomic"
)

// Sll is a scored linked list.
type Sll struct {
	root *Node
	len  uint64
}

// Node is a scored linked list node.
type Node struct {
	next  *Node
	prev  *Node
	list  *Sll
	Score uint64
	Value interface{}
}

// Next returns the next node in the *Sll.
func (n *Node) Next() *Node {
	if n.next != n.list.root {
		return n.next
	}

	return nil
}

// Prev returns the previous node in the *Sll.
func (n *Node) Prev() *Node {
	if n.prev != n.list.root {
		return n.prev
	}

	return nil
}

// Copy returns a copy of a *Node.
func (n *Node) Copy() *Node {
	return &Node{
		Score: n.Score,
		Value: n.Value,
	}
}

// New creates a new *Sll.
func New() *Sll {
	ll := &Sll{
		root: &Node{},
	}

	ll.root.next, ll.root.prev = ll.root, ll.root

	return ll
}

// nodeScoreList holds a slice of *Node
// for sorting by score.
type nodeScoreList []*Node

// Read returns a *Node Value and increments the score.
func (n *Node) Read() interface{} {
	atomic.AddUint64(&n.Score, 1)
	return n.Value
}

// nodeScoreList methods to satisfy the sort interface.

func (nsl nodeScoreList) Len() int {
	return len(nsl)
}

func (nsl nodeScoreList) Less(i, j int) bool {
	return atomic.LoadUint64(&nsl[i].Score) < atomic.LoadUint64(&nsl[j].Score)
}

func (nsl nodeScoreList) Swap(i, j int) {
	nsl[i], nsl[j] = nsl[j], nsl[i]
}

// Len returns the count of nodes in the *Sll.
func (ll *Sll) Len() uint {
	return uint(ll.len)
}

// Head returns the head *Node.
func (ll *Sll) Head() *Node {
	return ll.root.prev
}

// Tail returns the head *Node.
func (ll *Sll) Tail() *Node {
	return ll.root.next
}

// Copy returns a copy of a *Sll.
func (ll *Sll) Copy() *Sll {
	newll := New()

	for node := ll.Head(); node != nil; node = node.Prev() {
		c := node.Copy()
		newll.PushTailNode(c)
	}

	return newll
}

// HighScores takes an integer and returns the
// respective number of *Nodes with the higest scores
// sorted in ascending order.
func (ll *Sll) HighScores(k int) nodeScoreList {
	h := &MinHeap{}

	if ll.Len() == 0 {
		return nodeScoreList(*h)
	}

	heap.Init(h)

	// Add the first k nodes
	// to the heap. In a high scores selection,
	// we traverse from the head toward the
	// tail with the assumption that head nodes
	// are more probable to have higher
	// scores than tail nodes.
	node := ll.Head()
	for i := 0; i < k && node != nil; i++ {
		heap.Push(h, node)
		node = node.Prev()
	}

	var min = h.Peek().(*Node).Score

	// Iterate the rest of the list
	// while maintaining the current
	// heap len.
	for ; node != nil; node = node.Prev() {
		if node.Score > min {
			heap.Push(h, node)
			heap.Pop(h)
			min = h.Peek().(*Node).Score
		}
	}

	scores := nodeScoreList(*h)
	sort.Sort(scores)

	return scores
}

// LowScores takes an integer and returns the
// respective number of *Nodes with the lowest scores
// sorted in ascending order.
func (ll *Sll) LowScores(k int) nodeScoreList {
	h := &MaxHeap{}

	if ll.Len() == 0 {
		return nodeScoreList(*h)
	}

	// In a low scores selection,
	// we traverse from the tail toward the
	// head with the assumption that tail nodes
	// are more probable to have lower
	// scores than head nodes.
	node := ll.Tail()
	for i := 0; i < k && node != nil; i++ {
		heap.Push(h, node)
		node = node.Next()
	}

	var max = h.Peek().(*Node).Score

	// Iterate the rest of the list
	// while maintaining the current
	// heap len.
	for ; node != nil; node = node.Next() {
		if node.Score < max {
			heap.Push(h, node)
			heap.Pop(h)
			max = h.Peek().(*Node).Score
		}
	}

	scores := nodeScoreList(*h)
	sort.Sort(scores)

	return scores
}

// insertAt inserts node n
// at position at in the *Sll.
func insertAt(n, at *Node) {
	n.next = at.next
	at.next = n
	n.prev = at
	n.next.prev = n
}

// pull removes a *Node from
// its position in the *Sll, but
// doesn't remove the node from
// the nodeScoreList. This is used for
// repositioning nodes.
func pull(n *Node) {
	// Link next/prev nodes.
	n.next.prev, n.prev.next = n.prev, n.next
	// Remove references.
	n.next, n.prev = nil, nil
}

// MoveToHead takes a *Node and moves it
// to the front of the *Sll.
func (ll *Sll) MoveToHead(n *Node) {
	// Short-circuit if this
	// is already the head.
	if ll.root.prev == n {
		return
	}

	// Pull and move node.
	pull(n)
	insertAt(n, ll.root.prev)
}

// MoveToTail takes a *Node and moves it
// to the back of the *Sll.
func (ll *Sll) MoveToTail(n *Node) {
	// Short-circuit if this
	// is already the tail.
	if ll.root.next == n {
		return
	}

	// Pull and move node.
	pull(n)
	insertAt(n, ll.root)
}

// PushHead creates a *Node with value v
// at the head of the *Sll and returns a *Node.
func (ll *Sll) PushHead(v interface{}) *Node {
	n := &Node{
		Value: v,
		Score: 0,
		list:  ll,
	}

	atomic.AddUint64(&ll.len, 1)
	insertAt(n, ll.root.prev)

	return n
}

// PushTail creates a *Node with value v
// at the tail of the *Sll and returns a *Node.
func (ll *Sll) PushTail(v interface{}) *Node {
	n := &Node{
		Value: v,
		Score: 0,
		list:  ll,
	}

	atomic.AddUint64(&ll.len, 1)
	insertAt(n, ll.root)

	return n
}

// PushHeadNode pushes an existing node
// to the head of the *Sll.
func (ll *Sll) PushHeadNode(n *Node) {
	n.list = ll

	atomic.AddUint64(&ll.len, 1)
	insertAt(n, ll.root.prev)
}

// PushTailNode pushes an existing node
// to the tail of the *Sll.
func (ll *Sll) PushTailNode(n *Node) {
	n.list = ll

	// Increment len.
	atomic.AddUint64(&ll.len, 1)
	insertAt(n, ll.root)
}

// Remove removes a *Node from the *Sll.
func (ll *Sll) Remove(n *Node) {
	// Link next/prev nodes.
	n.next.prev, n.prev.next = n.prev, n.next

	// Remove references.
	n.next, n.prev = nil, nil

	// Decrement len.
	atomic.AddUint64(&ll.len, ^uint64(0))
}

// RemoveHead removes the current *Sll.head.
func (ll *Sll) RemoveHead() {
	ll.Remove(ll.root.prev)
}

// RemoveTail removes the current *Sll.tail.s
func (ll *Sll) RemoveTail() {
	ll.Remove(ll.root.next)
}
