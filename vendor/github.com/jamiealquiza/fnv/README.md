[![GoDoc](https://godoc.org/github.com/jamiealquiza/fnv?status.svg)](https://godoc.org/github.com/jamiealquiza/fnv)

# fnv

Package fnv implements allocation-free 32 and 64 bit FNV-1 hash variants.

```
% go test -bench=. -benchmem
BenchmarkHash32a-4      300000000                4.85 ns/op            0 B/op          0 allocs/op
BenchmarkHash32-4       300000000                4.78 ns/op            0 B/op          0 allocs/op
BenchmarkHash64a-4      300000000                6.01 ns/op            0 B/op          0 allocs/op
BenchmarkHash64-4       300000000                5.18 ns/op            0 B/op          0 allocs/op
```
