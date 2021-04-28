# fastrand

Package fastrand collects some high performance pseudo-random number generator algorithms.

Related websites:

1. https://lemire.me/blog/2016/06/27/a-fast-alternative-to-the-modulo-reduction/
1. https://lemire.me/blog/2016/06/30/fast-random-shuffling/
1. https://github.com/golang/exp/tree/master/rand
1. http://www.pcg-random.org/

Benchmark results:

```text
cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz

BenchmarkMathRandUint32-12                      327068841                3.548 ns/op
BenchmarkMathRandUint64-12                      243641504                4.895 ns/op
BenchmarkPCG32-12                               629150890                1.856 ns/op
BenchmarkPCG64-12                               525319977                2.269 ns/op

BenchmarkMathRand_Uint32n-12                    158754159                7.502 ns/op
BenchmarkMathRand_Uint64n-12                    60465216                17.91 ns/op
BenchmarkPCG32_Uint32n-12                       347265626                3.371 ns/op
BenchmarkPCG32_Uint32nRough-12                  637146206                1.853 ns/op
BenchmarkPCG64_Uint64n-12                       288755594                4.165 ns/op
BenchmarkPCG64_Uint64nRough-12                  305076949                3.932 ns/op

BenchmarkConcurrentRuntimeFastrand-12           1000000000               0.3648 ns/op
BenchmarkConcurrentMathRandUint32-12            18759075                56.39 ns/op
BenchmarkConcurrentMathRandUint64-12            18796676                58.11 ns/op
BenchmarkConcurrentPCG32-12                     1000000000               0.8837 ns/op
BenchmarkConcurrentPCG64-12                     1000000000               0.9993 ns/op
```
