// The MIT License (MIT)
//
// Copyright (c) 2016 Jamie Alquiza
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.
package main

import (
	"fmt"

	"github.com/jamiealquiza/bicache/sll"
)

func main() {
	s := sll.New(10)

	objects := []string{
		"one",
		"two",
		"three",
		"four",
		"five",
		"six",
		"seven",
		"eight",
		"nine",
		"ten",
	}

	fmt.Println("[ PushTail objects ]")
	for _, o := range objects {
		fmt.Println(o)
		_ = s.PushTail(o)
	}

	fmt.Printf("\n[ traverse from tail ]\n")
	node := s.Tail()
	for {
		fmt.Printf("%s -> ", node.Read())
		if node.Next != nil {
			node = node.Next
		} else {
			break
		}
	}

	fmt.Printf("\n\n[ traverse from head ]\n")
	node = s.Head()
	for {
		fmt.Printf("%s -> ", node.Read())
		if node.Prev != nil {
			node = node.Prev
		} else {
			break
		}
	}

	fmt.Printf("\n\n[ read tail 3x ]\n")
	tail := s.Tail()
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
	node = s.Tail()
	for {
		fmt.Printf("%s -> ", node.Read())
		if node.Next != nil {
			node = node.Next
		} else {
			break
		}
	}

	s.MoveToHead(s.Tail())

	fmt.Printf("\nNew: ")
	node = s.Tail()
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
	node = s.Tail()
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
	node = s.Tail()
	for {
		fmt.Printf("%s -> ", node.Read())
		if node.Next != nil {
			node = node.Next
		} else {
			break
		}
	}

	fmt.Printf("\n\n[ remove second from last node ]\n")
	node = s.Tail().Next
	s.Remove(node)

	node = s.Tail()
	for {
		fmt.Printf("%s -> ", node.Read())
		if node.Next != nil {
			node = node.Next
		} else {
			break
		}
	}

	fmt.Printf("\n\n[ read score list ]\n")
	for _, n := range s.HighScores(int(s.Len())) {
		fmt.Printf("%s:%d ", n.Read(), n.Score)
	}
	fmt.Println()
}
