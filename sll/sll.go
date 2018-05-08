package sll

import (
	"container/heap"
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

func (ll *Sll) heapSelect(h heap.Interface, k int) {
	if ll.Len() == 0 {
		return
	}

	heap.Init(h)
	// Add the first k nodes
	// to the heap.
	node := ll.Tail()
	for i := 0; i < k && node != nil; i++ {
		heap.Push(h, node)
		node = node.Next()
	}

	// Iterate the rest of the list
	// while maintaining the current
	// heap len.
	for node != nil {
		heap.Push(h, node)
		heap.Pop(h)
		node = node.Next()
	}
}

// HighScores takes an integer and returns the
// respective number of *Nodes with the higest scores
// sorted in ascending order.
func (ll *Sll) HighScores(k int) nodeScoreList {
	h := MinHeap{}
	ll.heapSelect(&h, k)

	scores := nodeScoreList(h)
	sort.Sort(scores)

	return scores
}

// LowScores takes an integer and returns the
// respective number of *Nodes with the lowest scores
// sorted in ascending order.
func (ll *Sll) LowScores(k int) nodeScoreList {
	h := MaxHeap{}
	ll.heapSelect(&h, k)

	scores := nodeScoreList(h)
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
	ll.removeFromScores(n)
}

// RemoveAsync removes a *Node from the *Sll
// and marks the node for removal. This is
// useful if a batch of many nodes are being
// removed, at the cost of the node score list
// being out of sync.
// The node score list must be updated
// with a subsequent call of the Sync() method
// once all desired nodes have been removed.
func (ll *Sll) RemoveAsync(n *Node) {
	// Link next/prev nodes.
	n.next.prev, n.prev.next = n.prev, n.next
	// Remove references.
	n.next, n.prev = nil, nil
	// Unset the parent list.
	// This is used as a removal marker
	// in the Sync() function.
	n.list = nil
}

// RemoveHead removes the current *Sll.head.
func (ll *Sll) RemoveHead() {
	ll.Remove(ll.root.prev)
}

// RemoveTail removes the current *Sll.tail.s
func (ll *Sll) RemoveTail() {
	ll.Remove(ll.root.next)
}

// RemoveHeadAsync removes the current *Sll.head
// using the RemoveAsync method.
func (ll *Sll) RemoveHeadAsync() {
	ll.RemoveAsync(ll.root.prev)
}

// RemoveTailAsync removes the current *Sll.tail
// using the RemoveAsync method.
func (ll *Sll) RemoveTailAsync() {
	ll.RemoveAsync(ll.root.next)
}

// Sync traverses the node score list
// and removes any marked for removal.
// This is typically called subsequent to
// many AsyncRemove ops.
func (ll *Sll) Sync() {
	// Prep an allocation-free filter slice.
	newScoreList := ll.scores[:0]

	// Traverse and exclude nodes
	// marked for removal.
	for n := range ll.scores {
		if ll.scores[n].list == ll {
			newScoreList = append(newScoreList, ll.scores[n])
		} else {
			// If a node is marked for removal,
			// nil the entry to avoid leaks.
			ll.scores[n] = nil
		}
	}

	// Update the ll.scores.
	ll.scores = newScoreList
}

// removeFromScores removes n from the nodeScoreList scores.
func (ll *Sll) removeFromScores(n *Node) {
	// Unrolling with 5 elements
	// has cut CPU-cached small element
	// slice search times in half. Needs further testing.
	// This will cause an out of bounds crash if the
	// element we're searching for somehow doesn't exist
	// (as a result of some other bug).
	var i int
	for p := 0; p < len(ll.scores); p += 5 {
		if ll.scores[p] == n {
			i = p
			break
		}
		if ll.scores[p+1] == n {
			i = p + 1
			break
		}
		if ll.scores[p+2] == n {
			i = p + 2
			break
		}
		if ll.scores[p+3] == n {
			i = p + 3
			break
		}
		if ll.scores[p+4] == n {
			i = p + 4
			break
		}
	}

	// Set the item to nil
	// to remove the reference in the
	// underlying slice array.
	ll.scores[i] = nil
	ll.scores = append(ll.scores[:i], ll.scores[i+1:]...)
}
