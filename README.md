## bloom

[![Go Reference](https://pkg.go.dev/badge/github.com/andrew-d/bloom.svg)](https://pkg.go.dev/github.com/andrew-d/bloom)

The `bloom` package is a zero-allocation implementation of a [Bloom Filter][bf]
in Go. It uses Go's [`hash/maphash`][maphash] package to generate hashes for
any [`comparable`][cmp] item with no allocations required.

[bf]: https://en.wikipedia.org/wiki/Bloom_filter
[maphash]: https://pkg.go.dev/hash/maphash
[cmp]: https://pkg.go.dev/builtin#comparable
