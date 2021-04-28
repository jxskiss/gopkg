# fastrand

Package fastrand collects some high performance pseudo-random number generator algorithms.

Related websites:

1. https://lemire.me/blog/2016/06/27/a-fast-alternative-to-the-modulo-reduction/
1. https://lemire.me/blog/2016/06/30/fast-random-shuffling/
1. https://github.com/golang/exp/tree/master/rand
1. http://www.pcg-random.org/

Benchmark results:

```text
cpu: Intel(R) Core(TM) i7-8750H CPU @ 2.20GHz
BenchmarkConcurrentRuntimeFastrand-12           1000000000               0.4169 ns/op
BenchmarkConcurrentMathRandUint32-12            14579678                69.62 ns/op
BenchmarkConcurrentMathRandUint64-12            14505070                71.39 ns/op
BenchmarkConcurrentPCG32-12                     1000000000               1.177 ns/op
BenchmarkConcurrentPCG64-12                     914565769                1.298 ns/op
```
