Package intintmap is a fork of https://github.com/brentp/intintmap/.
It implements a fast int64 key -> int64 value map.

Related articles:

1. [Implementing a world fastest Java int-to-int hash map](http://java-performance.info/implementing-world-fastest-java-int-to-int-hash-map/)
1. [Fibonacci Hashing: The Optimization that the World Forgot (or: a Better Alternative to Integer Modulo)](https://probablydance.com/2018/06/16/fibonacci-hashing-the-optimization-that-the-world-forgot-or-a-better-alternative-to-integer-modulo/)

It is 2-8X faster than the builtin map, benchmark:

```text
cpu: Intel(R) Core(TM) i7-8750H CPU @ 2.20GHz
BenchmarkIntIntMapFill-12                                     12          95630852 ns/op
BenchmarkStdMapFill-12                                         5         248595736 ns/op
BenchmarkIntIntMapGet10PercentHitRate-12                   11953             96504 ns/op
BenchmarkStdMapGet10PercentHitRate-12                      10083            107622 ns/op
BenchmarkIntIntMapGet100PercentHitRate-12                    870           1305507 ns/op
BenchmarkStdMapGet100PercentHitRate-12                       104          10678794 ns/op
BenchmarkIntIntMapGet_Size_1024_FillFactor_60-12       479573955                 2.189 ns/op
BenchmarkStdMapGet_Size_1024-12                         77867008                16.36 ns/op

cpu: Intel(R) Core(TM) i7-8750H CPU @ 2.20GHz
BenchmarkConcurrentStdMapGet_NoLock-12                  72221664                15.79 ns/op
BenchmarkConcurrentStdMapGet_RWMutex-12                  3255450               369.01 ns/op
BenchmarkConcurrentSyncMapGet-12                        27300724                44.59 ns/op
BenchmarkConcurrentCOWMapGet-12                        344179628                 3.487 ns/op
BenchmarkConcurrentSliceIndex-12                       908571164                 1.213 ns/op
```

Some notes:

```shell
# check inline cost information
go build -gcflags=-m=2 ./

# check bounds check elimination information
go build -gcflags="-d=ssa/check_bce/debug=1" ./

# check assembly output
go tool compile -S ./intintmap.go
```
