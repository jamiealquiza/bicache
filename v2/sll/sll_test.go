package sll_test

import (
	"math/rand"
	"testing"
	// "fmt"

	"github.com/jamiealquiza/bicache/v2/sll"
)

func TestHead(t *testing.T) {
	s := sll.New()

	node := s.PushHead("value")
	if s.Head() != node {
		t.Error("Unexpected head node")
	}
}

func TestTail(t *testing.T) {
	s := sll.New()

	node := s.PushTail("value")
	if s.Tail() != node {
		t.Error("Unexpected tail node")
	}
}

func TestRead(t *testing.T) {
	s := sll.New()

	s.PushHead("value")
	if s.Head().Read() != "value" {
		t.Error("Read method failed")
	}
}

func TestPushHead(t *testing.T) {
	s := sll.New()

	s.PushHead("value")
	if s.Head().Read() != "value" {
		t.Errorf(`Expected value "value", got "%s"`, s.Head().Read())
	}
}

func TestPushTail(t *testing.T) {
	s := sll.New()

	s.PushTail("value")
	if s.Tail().Read() != "value" {
		t.Errorf(`Expected value "value", got "%s"`, s.Tail().Read())
	}
}

func TestNext(t *testing.T) {
	s := sll.New()

	firstVal := "first"
	secondVal := "second"

	first := s.PushTail(firstVal)
	second := s.PushTail(secondVal)

	if second.Next() != first {
		t.Errorf("Expected node with value %s next, got %s", firstVal, second.Next().Read())
	}
}

func TestPrev(t *testing.T) {
	s := sll.New()

	firstVal := "first"
	secondVal := "second"

	first := s.PushTail(firstVal)
	second := s.PushTail(secondVal)

	if first.Prev() != second {
		t.Errorf("Expected node with value %s next, got %s", firstVal, second.Next().Read())
	}
}

func TestLen(t *testing.T) {
	s := sll.New()

	for i := 0; i < 5; i++ {
		s.PushTail(i)
	}

	if s.Len() != 5 {
		t.Errorf("Expected len 5, got %d", s.Len())
	}
}

func TestHighScores(t *testing.T) {
	s := sll.New()

	nodes := map[int]*sll.Node{}

	for i := 0; i < 5; i++ {
		nodes[i] = s.PushTail(i)
	}

	nodes[3].Read()
	nodes[3].Read()
	nodes[3].Read()

	nodes[4].Read()
	nodes[4].Read()

	// Should result in [0,4,3] with read scores
	// 0,2,3 respectively.

	scores := s.HighScores(3)

	// for node := range nodes {
	// 	fmt.Printf("node %d: %d\n", node, nodes[node].Score)
	// }
	//
	// fmt.Println("-")
	//
	// for _, node := range scores {
	// 	fmt.Printf("node %d: %d\n", node.Value, node.Score)
	// }

	if scores[0] != nodes[2] {
		t.Errorf("Expected scores position 0 node with value 2, got %d", scores[0].Read())
	}

	if scores[1] != nodes[4] {
		t.Errorf("Expected scores position 1 node with value 4, got %d", scores[1].Read())
	}

	if scores[2] != nodes[3] {
		t.Errorf("Expected scores position 2 node with value 3, got %d", scores[2].Read())
	}
}

func TestLowScores(t *testing.T) {
	s := sll.New()

	nodes := map[int]*sll.Node{}

	for i := 0; i < 3; i++ {
		nodes[i] = s.PushTail(i)
	}

	nodes[1].Read()
	nodes[1].Read()
	nodes[1].Read()

	nodes[0].Read()
	nodes[0].Read()

	// Should result in [2, 0, 1]
	// with read scores of 0, 2, 3 respectively.
	scores := s.LowScores(3)

	/*
		for node := range nodes {
			fmt.Printf("node %d: %d\n", node, nodes[node].Score)
		}

		fmt.Println("-")

		for _, node := range scores {
			fmt.Printf("node %d: %d\n", node.Value, node.Score)
		}
	*/

	if scores[0] != nodes[2] {
		t.Errorf("Expected scores position 0 node with value 2, got %d", scores[2].Read())
	}

	if scores[1] != nodes[0] {
		t.Errorf("Expected scores position 1 node with value 4, got %d", scores[0].Read())
	}

	if scores[2] != nodes[1] {
		t.Errorf("Expected scores position 2 node with value 3, got %d", scores[1].Read())
	}
}

func benchmarkHeapScores(b *testing.B, l int) {
	b.N = 1
	b.StopTimer()

	// Create/populate an sll.
	s := sll.New()
	for i := 0; i < l; i++ {
		s.PushTail(i)
	}

	// Perform 3*sll.Len() reads
	// on random nodes to produce
	// random node score counts.
	node := s.Tail()
	for i := 0; i < 3*l; i++ {
		node.Read()
		for j := 0; j < rand.Intn(10); j++ {
			node = node.Next()
			if node == nil {
				node = s.Tail()
			}
		}
	}

	b.ResetTimer()
	b.StartTimer()

	for n := 0; n < b.N; n++ {
		// Call HighScores for
		// 1/20th the sll len.
		s.HighScores(l / 20)
	}
}

func BenchmarkHeapScores200K(b *testing.B) { benchmarkHeapScores(b, 200000) }
func BenchmarkHeapScores2M(b *testing.B)   { benchmarkHeapScores(b, 2000000) }

func TestScoresEmpty(t *testing.T) {
	s := sll.New()

	hScores := s.HighScores(5)
	lScores := s.LowScores(5)

	// We don't really care about the
	// len; if it's really broken,
	// we'd probably have crashed
	// by now.

	if len(hScores) != 0 {
		t.Errorf("Expected scores len of 0, got %d", len(hScores))
	}

	if len(lScores) != 0 {
		t.Errorf("Expected scores len of 0, got %d", len(lScores))
	}
}

func TestMoveToHead(t *testing.T) {
	s := sll.New()

	for i := 0; i < 10; i++ {
		s.PushTail(i)
	}

	node := s.Tail().Next().Next()
	s.MoveToHead(node)

	// Check head method.
	if s.Head() != node {
		t.Errorf(`Expected node with value "%d" at head, got "%d"`,
			node.Read(), s.Head().Read())
	}

	// Check total order.
	expected := []int{9, 8, 6, 5, 4, 3, 2, 1, 0, 7}
	var i int
	for n := s.Tail(); n != nil; n = n.Next() {
		if n.Read() != expected[i] {
			t.Errorf(`Expected node with vallue "%d", got "%d"`, expected[i], n.Read())
		}
		i++
	}
}

func TestMoveToTail(t *testing.T) {
	s := sll.New()

	for i := 0; i < 10; i++ {
		s.PushTail(i)
	}

	node := s.Tail().Next().Next()
	s.MoveToTail(node)

	// Check tail method.
	if s.Tail() != node {
		t.Errorf(`Expected node with value "%d" at tail, got "%d"`,
			node.Read(), s.Tail().Read())
	}

	// Check total order.
	expected := []int{7, 9, 8, 6, 5, 4, 3, 2, 1, 0}
	var i int
	for n := s.Tail(); n != nil; n = n.Next() {
		if n.Read() != expected[i] {
			t.Errorf(`Expected node with value "%d", got "%d"`, expected[i], n.Read())
		}
		i++
	}
}

func TestPushHeadNode(t *testing.T) {
	s1 := sll.New()
	s2 := sll.New()

	s1.PushTail("target")
	node := s1.Tail()
	s1.Remove(node)

	s2.PushHead("value")
	s2.PushHead("value")
	s2.PushHeadNode(node)

	// Check from the head.
	if s2.Head() != node {
		t.Errorf(`Expected node with value "target", got "%s"`, s2.Head().Read())
	}

	// Ensure the links are correct.
	if s2.Tail().Next().Next() != node {
		t.Errorf(`Expected node with value "target", got "%s"`, s2.Tail().Next().Next().Read())
	}
}

func TestPushTailNode(t *testing.T) {
	s1 := sll.New()
	s2 := sll.New()

	s1.PushTail("target")
	node := s1.Tail()
	s1.Remove(node)

	s2.PushHead("value")
	s2.PushHead("value")
	s2.PushTailNode(node)

	// Check from the head.
	if s2.Tail() != node {
		t.Errorf(`Expected node with value "target", got "%s"`, s2.Tail().Read())
	}

	// Ensure the links are correct.
	if s2.Head().Prev().Prev() != node {
		t.Errorf(`Expected node with value "target", got "%s"`, s2.Head().Prev().Prev().Read())
	}
}

func TestRemove(t *testing.T) {
	s := sll.New()

	nodes := map[int]*sll.Node{}

	for i := 0; i < 3; i++ {
		nodes[i] = s.PushTail(i)
	}

	s.Remove(nodes[1])

	if s.Tail().Next().Read() != 0 {
		t.Errorf(`Expected node with value "0", got "%d"`, s.Tail().Next().Read())
	}

	scores := s.HighScores(3)

	if len(scores) != 2 {
		t.Errorf("Expected scores len 2, got %d", len(scores))
	}

	// This effectively tests the unexported
	// removeFromScores method.
	if scores[0] != s.Tail() {
		t.Error("Unexpected node in scores position 0")
	}

	if scores[1] != s.Tail().Next() {
		t.Error("Unexpected node in scores position 1")
	}
}

func TestRemoveHead(t *testing.T) {
	s := sll.New()

	s.PushTail("value")
	target := s.PushTail("value")
	s.PushTail("value")

	s.RemoveHead()

	if s.Head() != target {
		t.Error("Unexpected head node")
	}
}

func TestRemoveTail(t *testing.T) {
	s := sll.New()

	s.PushTail("value")
	target := s.PushTail("value")
	s.PushTail("value")

	s.RemoveTail()

	if s.Tail() != target {
		t.Error("Unexpected tail node")
	}
}
