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

	fmt.Println("\n[ PushTail objects ]\n")
	for _, o := range objects {
		fmt.Println(o)
		s.PushTail(o)
	}

	fmt.Printf("\n\n[ traverse from tail ]\n")
	next := s.Tail
	for {
		fmt.Println(next.Read())
		if next.Next != nil {
			next = next.Next
		} else {
			break
		}
	}

	fmt.Printf("\n\n[ traverse from head ]\n")
	next = s.Head
	for {
		fmt.Println(next.Read())
		if next.Prev != nil {
			next = next.Prev
		} else {
			break
		}
	}

	fmt.Printf("\n\n[ read tail ]\n")
	tail := s.Tail
	fmt.Println(tail.Read())
	fmt.Println(tail.Read())
	fmt.Println(tail.Read())

	fmt.Printf("\n\n[ read top 2 scores ]\n")
	top2 := s.HighScores(2)
	for _, o := range top2 {
		fmt.Printf("Value:%s Score:%d\n",
			o.Value, o.Score)
	}
}