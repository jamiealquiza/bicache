package sll

// MinHeap implements a min-heap heap.Interface.
type MinHeap []*Node

func (mh MinHeap) Len() int { return len(mh) }

func (mh MinHeap) Less(i, j int) bool {
	return mh[i].Score < mh[j].Score
}

func (mh MinHeap) Swap(i, j int) {
	mh[i], mh[j] = mh[j], mh[i]
}

// Push adds an item to the heap.
func (mh *MinHeap) Push(x interface{}) {
	item := x.(*Node)
	*mh = append(*mh, item)
}

// Pop removes and returns the root node from the heap.
func (mh *MinHeap) Pop() interface{} {
	old := *mh
	n := len(old)
	item := old[n-1]
	*mh = old[0 : n-1]
	return item
}

// Peek returns the root node from the heap.
func (mh *MinHeap) Peek() interface{} {
	s := *mh
	return s[0]
}

// MaxHeap implements a max-heap heap.Interface.
type MaxHeap []*Node

func (mh MaxHeap) Len() int { return len(mh) }

func (mh MaxHeap) Less(i, j int) bool {
	return mh[i].Score > mh[j].Score
}

func (mh MaxHeap) Swap(i, j int) {
	mh[i], mh[j] = mh[j], mh[i]
}

// Push adds an item to the heap.
func (mh *MaxHeap) Push(x interface{}) {
	item := x.(*Node)
	*mh = append(*mh, item)
}

// Pop removes and returns the root node from the heap.
func (mh *MaxHeap) Pop() interface{} {
	old := *mh
	n := len(old)
	item := old[n-1]
	*mh = old[0 : n-1]
	return item
}

// Peek returns the root node from the heap.
func (mh *MaxHeap) Peek() interface{} {
	s := *mh
	return s[0]
}
