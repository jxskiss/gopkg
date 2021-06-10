# fastrand

## Runtime fastrand

Package fastrand exported a set of pseudo-random generator methods wrapped around the fastrand
function from the Go runtime. There is a generator per-M (physical thread), thus it doesn't
need to do synchronization when generate random sequence, which makes it very scalable.

## PCG

Except the exported functions wrapped around runtime fastrand,
this package also include a PCG generator implementation,
which is forked from https://github.com/golang/exp/tree/master/rand.
Compared to the PRNG generator used by `math/rand`, it uses a very small state,
and it's cheap to create many generators.

In case you want a separate generator for each use-case, you may use the PCG generator. <br>
Or you may want a reproducible sequence, then you can create a PCG generator and seed the
generator with desired state, it will give you the same sequence with the same seed.

## Related resources

1. https://github.com/golang/exp/tree/master/rand
1. http://www.pcg-random.org/
1. https://lemire.me/blog/2016/06/27/a-fast-alternative-to-the-modulo-reduction/
1. https://lemire.me/blog/2016/06/30/fast-random-shuffling/

## Benchmark results

```text
cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz

BenchmarkRuntimeFastrand_Uint32-12                      436941264                2.722 ns/op
BenchmarkMathRand_Uint32-12                             267724545                4.163 ns/op
BenchmarkPCG64_Uint32-12                                305560921                3.745 ns/op

BenchmarkRuntimeFastrand_Int63n-12                      155718987                7.527 ns/op
BenchmarkMathRand_Int63n-12                             63910872                21.07 ns/op
BenchmarkPCG64_Int63n-12                                237976500                5.488 ns/op

BenchmarkConcurrentRuntimeFastrand_Uint32-12            1000000000               0.3973 ns/op
BenchmarkConcurrentRuntimeFastrand_Uint64-12            1000000000               0.9080 ns/op
BenchmarkConcurrentMathRand_Uint32-12                   16165498                67.40 ns/op
BenchmarkConcurrentMathRand_Uint64-12                   19130859                66.36 ns/op
```
