# gopkg

This repository collects many frequently used small packages, it is designed
to have reasonable trade-off between generic, performance and API friendliness.
It helps to make life easier!

## Status

This v1 branch is the legacy code which is released as versions v0.x.x.
This branch is not actively maintained, there shall be no further changes to the v1 APIs.

A new v2 branch is being built, which will be released as "github.com/jxskiss/gopkg/v2".
v2 requires Go 1.18+, users are recommended to upgrade to v2 when the generics feature is stable.

## Packages

1. `confr` provides a simple but powerful configuration loader.

1. `easy` contains many handy utilities as a complementary to the standard library,
   it really makes life easier.

1. `errcode` provides a registry to manage error codes and messages.

1. `forceexport` force exports private types and functions of other packages. As you might
   expect, this package is unsafe and fragile and shouldn't be used in production, it is
   mainly for testing purpose.

1. `gemap` contains some utilities to manipulate map data structure.

1. `json` provides a drop-in replacement of `encoding/json` and extended features.
   When compiled with tag `gojson`, it uses the `goccy/go-json` in underlying, which has
   much better performance then `encoding/json` and many other third-party libraries.
   Check json/README.md for detailed introduction.

1. `lru` is a high performance implementation of the LRU cache, it features pre-allocation,
   item expiration, friendly and type-safe APIs for commonly used key types.
   It also provides a sharded version for heavy lock contention use-case.

1. `ptr` provides small functions to work with pointer types, such as copying value as pointer,
   converting integer to string pointer, or dereference pointer which may be nil, etc.
   It helps to reduce duplicate code and makes code clearer and simpler.

1. `reflectx` contains many utilities to work with reflection, providing convenient APIs
   or better performance.

1. `retry` implements frequently used strategies and options to do retry when error
   occurs, with various hooks and circuit breaker to protect system overload.

1. `rthash` exposes the various hash functions in runtime package.

1. `set` provides a generic set data structure, and some most frequently used concrete types.

1. `sqlutil` provides utilities to work with relational database ORM libraries,
   it keeps simple and can work with `database/sql` and many ORM libraries.

1. `structtag` contains some handy utilities to work with struct tags.

1. `strutil` provides utilities to work with string data as supplement to the standard
   libraries `strings` and `unicode/utf8`.

1. `zlog` provides a common logger interface which is implemented by
   many logging libraries, and some opinionated logging facilities wrapped around
   [zap](https://github.com/uber-go/zap).

See https://pkg.go.dev/github.com/jxskiss/gopkg for detailed online docs.
