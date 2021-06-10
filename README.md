# gopkg

This repository collects many frequently used small packages, it is designed
to have reasonable trade-off between generic, performance and API friendliness.
It helps to make life easier!

## Code layout

Packages under `exp` directory are considered experimental and unreliable,
the API may change arbitrarily, or even disappear some day without any notice.
They do not follow the Go 1 compatibility promise.
When a package is tested enough, and the API is considered stable, it may be promoted
as "stable" and moved under the top directory.

Packages under the top directory are considered "stable", they try to follow the
Go 1 compatibility promise and the Semantic Versioning spec.
Packages under the top directory may depend on `exp` packages but must keep
reasonable API compatibility and versioning.

## Packages

1. `bbp` provides efficient byte buffer pools with anti-memory-waste protection.

1. `easy` contains many handy utilities as a complementary to the standard library,
   it really makes life easier.

1. `encrypt` contains some encryption utilities.

1. `lru` is a high performance implementation of the LRU cache, it features pre-allocation,
   item expiration, friendly and type-safe APIs for commonly used key types.
   It also provides a sharded version for heavy lock contention use-case.

1. `ptr` provides small functions to work with pointer types, such as copying value as pointer,
   converting integer to string pointer, or dereference pointer which may be nil, etc.
   It helps to reduce duplicate code and makes code clearer and simpler.

1. `retry` implements frequently used strategies and options to do retry when error
   occurs, with various hooks and circuit breaker to protect system overload.

1. `rthash` exposes the various hash functions in runtime package.

1. `set` provides set data structure for 

1. `sqlutil` provides utilities to work with relational database ORM libraries,
   it keeps simple and can work with `database/sql` and many ORM libraries.

1. `strutil` provides utilities to work with string data as supplement to the standard
   libraries `strings` and `unicode/utf8`.

See https://pkg.go.dev/github.com/jxskiss/gopkg for detailed online docs.

## TODO

[ ] More tests and documents to ship a stable release.
