[![GoDoc](https://godoc.org/github.com/jamiealquiza/bicache/sll?status.svg)](https://godoc.org/github.com/jamiealquiza/bicache/sll)


# Sll
A scored linked list. Sll implements a pointer-based doubly linked list with the addition of methods to fetch nodes by score (high or low) and arbitrarily move nodes between lists. A node score is incremented with each `Read()` method called while retrieving the node's value.

- See [GoDoc](https://godoc.org/github.com/jamiealquiza/bicache/sll) for reference.
- See [`sll-example`](./sll-example) for example usage.
