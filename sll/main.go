package sll

import (
	"sort"
	"sync/atomic"
)

// Sll is a scored linked list.
type Sll struct {
	root   *Node
	scores nodeScoreList
}

// Node is a scored linked list node.
type Node struct {
	next  *Node
	prev  *Node
	list  *Sll
	Score uint64
	Value interface{}
	// Might add a create field for TTL.
}

func (n *Node) Next() *Node {
	if n.next != n.list.root {
		return n.next
	}

	return nil
}

func (n *Node) Prev() *Node {
	if n.prev != n.list.root {
		return n.prev
	}

	return nil
}

// New creates a new *Sll. New takes an
// integer length to pre-allocate a nodeScoreList
// of capacity l. This reduces append latencies if
// many elements are inserted into a new list.
func New(l int) *Sll {
	ll := &Sll{
		root:   &Node{},
		scores: make(nodeScoreList, 0, l),
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
	return uint(len(ll.scores))
}

// Head returns the head *Node.
func (ll *Sll) Head() *Node {
	return ll.root.prev
}

// Tail returns the head *Node.
func (ll *Sll) Tail() *Node {
	return ll.root.next
}

// HighScores takes an integer and returns the
// respective number of *Nodes with the higest scores
// sorted in ascending order. The last element will
// be the node with the highest score.
func (ll *Sll) HighScores(r int) nodeScoreList {
	sort.Sort(ll.scores)
	// Return what's available
	// if more is being requested
	// than exists.
	if r > len(ll.scores) {
		scores := make(nodeScoreList, len(ll.scores))
		copy(scores, ll.scores)
		return scores
	}

	// We return a copy because the
	// underlying array order will
	// possibly change.
	scores := make(nodeScoreList, r)
	copy(scores, ll.scores[len(ll.scores)-r:])

	return scores
}

// LowScores takes an integer and returns the
// respective number of *Nodes with the lowest scores
// sorted in ascending order. The first element will
// be the node with the lowest score.
func (ll *Sll) LowScores(r int) nodeScoreList {
	sort.Sort(ll.scores)
	// Return what's available
	// if more is being requested
	// than exists.
	if r > len(ll.scores) {
		scores := make(nodeScoreList, len(ll.scores))
		copy(scores, ll.scores)
		return scores
	}

	// We return a copy because the
	// underlying array order will
	// possibly change.
	scores := make(nodeScoreList, r)
	copy(scores, ll.scores[:r])

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

	// Add to scores and insert.
	ll.scores = append(ll.scores, n)
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

	// Add to scores and insert.
	ll.scores = append(ll.scores, n)
	insertAt(n, ll.root)

	return n
}

// PushHeadNode pushes an existing node
// to the head of the *Sll.
func (ll *Sll) PushHeadNode(n *Node) {
	n.list = ll

	// Add to scores and insert.
	ll.scores = append(ll.scores, n)
	insertAt(n, ll.root.prev)
}

// PushTailNode pushes an existing node
// to the tail of the *Sll.
func (ll *Sll) PushTailNode(n *Node) {
	n.list = ll

	// Add to scores and insert.
	ll.scores = append(ll.scores, n)
	insertAt(n, ll.root)
}

// Remove removes a *Node from the *Sll.
func (ll *Sll) Remove(n *Node) {
	// Link next/prev nodes.
	n.next.prev, n.prev.next = n.prev, n.next
	// Remove references.
	n.next, n.prev = nil, nil
	//Update scores.
	ll.scores = removeFromScores(ll.scores, n)
}

// RemoveHead removes the current *Sll.head.
func (ll *Sll) RemoveHead() {
	ll.Remove(ll.root.prev)
}

// RemoveTail removes the current *Sll.tail.s
func (ll *Sll) RemoveTail() {
	ll.Remove(ll.root.next)
}

// removeFromScores removes n from the nodeScoreList scores.
func removeFromScores(scores nodeScoreList, n *Node) nodeScoreList {
	// Binary search was demonstrating
	// incredible latencies (even excluding sort time).
	// Disabled in favor of an unrolled linear search for now.

	// Binary search also doesn't work when all values (scores)
	// are the same, even though the node is certain to exist.

	// Unrolling with 5 elements
	// has cut CPU-cached small element
	// slice search times in half. Needs further testing.
	// This will cause an out of bounds crash if the
	// element we're searching for somehow doesn't exist
	// (as a result of some other bug).
	var i int
	for p := 0; p < len(scores); p += 5 {
		if scores[p] == n {
			i = p
			break
		}
		if scores[p+1] == n {
			i = p + 1
			break
		}
		if scores[p+2] == n {
			i = p + 2
			break
		}
		if scores[p+3] == n {
			i = p + 3
			break
		}
		if scores[p+4] == n {
			i = p + 4
			break
		}
	}

	newScoreList := make(nodeScoreList, len(scores)-1)

	if i == len(scores)-1 {
		// If the index is at the tail,
		// we just exclude the last element.
		copy(newScoreList, scores)

	} else {
		copy(newScoreList, scores[:i])
		copy(newScoreList[i:], scores[i+1:])
	}

	return newScoreList
}
