package main

import (
	"fmt"

	"github.com/jamiealquiza/bicache/sll"
)

func main() {
	s := sll.New()

	objects := []string{
		"one",
		"two",
		"three",
		"four",
		"five",
	}

	fmt.Println("[ PushTail objects ]")
	for _, o := range objects {
		fmt.Println(o)
		s.PushTail(o)
	}

	fmt.Printf("\n[ traverse from tail ]\n")
	node := s.Tail
	for {
		fmt.Printf("%s -> ", node.Read())
		if node.Next != nil {
			node = node.Next
		} else {
			break
		}
	}

	fmt.Printf("\n\n[ traverse from head ]\n")
	node = s.Head
	for {
		fmt.Printf("%s -> ", node.Read())
		if node.Prev != nil {
			node = node.Prev
		} else {
			break
		}
	}

	fmt.Printf("\n\n[ read tail 3x ]\n")
	tail := s.Tail
	fmt.Println(tail.Read())
	fmt.Println(tail.Read())
	fmt.Println(tail.Read())

	fmt.Printf("\n[ read top 2 scores ]\n")
	top2 := s.HighScores(2)
	for _, o := range top2 {
		fmt.Printf("Value:%s Score:%d\n",
			o.Value, o.Score)
	}

	fmt.Printf("\n[ move tail to head ]\n")
	fmt.Printf("Current: ")
	node = s.Tail
	for {
		fmt.Printf("%s -> ", node.Read())
		if node.Next != nil {
			node = node.Next
		} else {
			break
		}
	}

	s.MoveToHead(s.Tail)

	fmt.Printf("\nNew: ")
	node = s.Tail
	for {
		fmt.Printf("%s -> ", node.Read())
		if node.Next != nil {
			node = node.Next
		} else {
			break
		}
	}

	fmt.Printf("\n\n[ remove head, tail ]\n")
	fmt.Printf("Current: ")
	node = s.Tail
	for {
		fmt.Printf("%s -> ", node.Read())
		if node.Next != nil {
			node = node.Next
		} else {
			break
		}
	}

	s.RemoveHead()
	s.RemoveTail()

	fmt.Printf("\nNew: ")
	node = s.Tail
	for {
		fmt.Printf("%s -> ", node.Read())
		if node.Next != nil {
			node = node.Next
		} else {
			break
		}
	}

	fmt.Printf("\n\n[ remove middle node ]\n")
	node = s.Tail.Next
	s.Remove(node)

	node = s.Tail
	for {
		fmt.Printf("%s -> ", node.Read())
		if node.Next != nil {
			node = node.Next
		} else {
			break
		}
	}

	fmt.Println()
}
