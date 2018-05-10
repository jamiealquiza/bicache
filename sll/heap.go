package sll

// A MinHeap implements heap.Interface and holds Nodes.
type MinHeap []*Node

func (mh MinHeap) Len() int { return len(mh) }

func (mh MinHeap) Less(i, j int) bool {
	return mh[i].Score < mh[j].Score
}

func (mh MinHeap) Swap(i, j int) {
	mh[i], mh[j] = mh[j], mh[i]
}

func (mh *MinHeap) Push(x interface{}) {
	item := x.(*Node)
	*mh = append(*mh, item)
}

func (mh *MinHeap) Pop() interface{} {
	old := *mh
	n := len(old)
	item := old[n-1]
	*mh = old[0 : n-1]
	return item
}

func (mh *MinHeap) Peek() interface{} {
	s := *mh
	return s[0]
}

type MaxHeap []*Node

func (mh MaxHeap) Len() int { return len(mh) }

func (mh MaxHeap) Less(i, j int) bool {
	return mh[i].Score > mh[j].Score
}

func (mh MaxHeap) Swap(i, j int) {
	mh[i], mh[j] = mh[j], mh[i]
}

func (mh *MaxHeap) Push(x interface{}) {
	item := x.(*Node)
	*mh = append(*mh, item)
}

func (mh *MaxHeap) Pop() interface{} {
	old := *mh
	n := len(old)
	item := old[n-1]
	*mh = old[0 : n-1]
	return item
}

func (mh *MaxHeap) Peek() interface{} {
	s := *mh
	return s[0]
}
