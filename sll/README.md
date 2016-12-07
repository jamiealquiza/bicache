[![GoDoc](https://godoc.org/github.com/jamiealquiza/bicache/sll?status.svg)](https://godoc.org/github.com/jamiealquiza/bicache/sll)


# Sll
A scored linked list. Sll implements a pointer-based doubly linked list with the addition of methods to fetch nodes by score (high or low). A node score is incremented with each `Read()` method called while retrieving the node's value.

Sll is somewhat specialized. The added overhead of scoring would possibly make Sll a poor choice in the case the functionality is not required; Sll simply bakes in some accounting overhead that would otherwise exist in external data structures.

- See [GoDoc](https://godoc.org/github.com/jamiealquiza/bicache/sll) for reference.
- See [`sll-example`](./sll-example) for example usage.
