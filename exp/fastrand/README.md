# fastrand

Package fastrand collects some high quality pseudo-random number generator algorithms.

Benchmark results:

```text
goos: darwin
goarch: amd64
pkg: github.com/jxskiss/gopkg/exp/fastrand
cpu: Intel(R) Core(TM) i7-8750H CPU @ 2.20GHz
BenchmarkMathRand-12            81305268                14.08 ns/op            0 B/op          0 allocs/op
BenchmarkRuntimeFastrand-12     460976776                2.239 ns/op           0 B/op          0 allocs/op
BenchmarkPCG32-12               202323628                6.097 ns/op           0 B/op          0 allocs/op
BenchmarkPCG64-12               171747643                7.068 ns/op           0 B/op          0 allocs/op
```
