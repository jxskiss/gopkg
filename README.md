# gopkg

[![GoDoc](https://img.shields.io/badge/api-Godoc-blue.svg)][godoc]
[![Go Report Card](https://goreportcard.com/badge/github.com/jxskiss/gopkg/v2)][goreport]
[![Issues](https://img.shields.io/github/issues/jxskiss/gopkg.svg)][issues]
[![GitHub release](http://img.shields.io/github/release/jxskiss/gopkg.svg)][release]
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg)][license]

[godoc]: https://pkg.go.dev/github.com/jxskiss/gopkg/v2
[goreport]: https://goreportcard.com/report/github.com/jxskiss/gopkg/v2
[issues]: https://github.com/jxskiss/gopkg/issues
[release]: https://github.com/jxskiss/gopkg/releases
[license]: https://github.com/jxskiss/gopkg/blob/master/LICENSE

This repository contains many frequently used small packages, it is designed
to have reasonable trade-off between generic, performance and API friendliness.
It helps to make life easier!

## Status

The master branch (v2) requires Go 1.18+, it uses the generics feature available in Go 1.18+.
For version compatibility policy, please refer to the following docs.

The v1 branch contains the legacy code.
there shall be no further changes to the v1 APIs.

## Code layout

Packages under `exp` directory are considered experimental and unreliable,
the API may change arbitrarily, or even disappear some day without any notice.
They do not follow the Go 1 compatibility promise.
When a package is tested enough, and the API is considered stable, it may be promoted
as "stable" and moved under the top directory.

Packages under the top directory are considered stable, they follow the
Go 1 compatibility promise and the Semantic Versioning spec.
Packages under the top directory may depend on `exp` packages but must keep
reasonable API compatibility and versioning.

## Packages

1. `collection/set` provides generic set data structures.

1. `confr` provides a simple but powerful configuration loader.

1. `easy` contains many handy utilities as a complementary to the standard library,
   it really makes life easier.

1. `easy/ezdbg` provides easy to use utilities which helps to do quick development.

1. `easy/ezhttp` contains some utilities to work with HTTP requests.

1. `easy/ezmap` contains some utilities to manipulate map data structure.

1. `encrypt` contains some encryption utilities.

1. `exp/kvutil` contains utilities to work with key-value cache.

1. `exp/mselect` can run many channel receiving operations simultaneously,
   tasks can be added dynamically.

1. `exp/singleflight` contains an async cache which can be used to fetch and update the
   latest data periodically and supports expiring a key if it's unused for a period.
   To be easy to use, it also re-exports `Group` and `Result` as alias names from package
   "golang.org/x/sync/singleflight".

1. `infra/errcode` provides error code registry to manage error codes and messages.
   Standardized error codes greatly helps in large deployment of microservice.

1. `infra/logid` provides log ID generators to help service observability, such as
   tracing, metrics, etc.

1. `perf/bbp` provides efficient byte buffer pools with anti-memory-waste protection.

1. `perf/fastrand` exported a set of pseudo-random generator methods wrapped around the fastrand
   function from the Go runtime. There is a generator per-M (physical thread), thus it doesn't
   need to do synchronization when generate random sequence, which makes it very scalable.

1. `perf/gopool` is a fork of `github.com/bytedance/gopkg/util/gopool`
   with opinionated changes.
   It's a high-performance goroutine pool which aims to reuse goroutines
   and limit the number of goroutines.

1. `perf/json` provides a drop-in replacement of `encoding/json` and extended features.
   By default, it uses [bytedance/sonic](https://github.com/bytedance/sonic) in underlying,
   which has better performance than `encoding/json` and many other third-party libraries.
   Check json/README.md for detailed introduction.

1. `perf/lru` is a high performance implementation of the LRU cache, it features pre-allocation,
   item expiration, friendly and type-safe APIs for commonly used key types.
   It also provides a sharded version for heavy lock contention use-case.

1. `perf/rthash` exposes the various hash functions in runtime package.

1. `perf/syncx` contains some synchronization primitives as an addition to the standard library,
   for better performance or alternative APIs.

1. `unsafe/forceexport` force exports private types and functions of other packages. As you might
   expect, this package is unsafe and fragile and shouldn't be used in production, it is
   mainly for testing purpose.

1. `unsafe/monkey` provides a method to do testing using the monkey-patch pattern.

1. `unsafe/reflectx` contains many utilities to work with reflection, providing convenient APIs
   or better performance.

1. `utils/ptr` provides small functions to work with pointer types, such as copying value as pointer,
   converting integer to string pointer, or dereference pointer which may be nil, etc.
   It helps to reduce duplicate code and makes code clearer and simpler.

1. `utils/retry` implements frequently used strategies and options to do retry when error
   occurs, with various hooks and circuit breaker to protect system overload.

1. `utils/sqlutil` provides utilities to work with relational database ORM libraries,
   it keeps simple and can work with `database/sql` and many ORM libraries.

1. `utils/structtag` contains some handy utilities to work with struct tags.

1. `utils/strutil` provides utilities to work with string data as supplement to the standard
   libraries `strings` and `unicode/utf8`.

1. `utils/vdutil` helps to do data validation.

1. `zlog` provides a common logger interface which is implemented by
   many logging libraries, and some opinionated logging facilities wrapped around
   [zap](https://github.com/uber-go/zap).

See [GoDoc][godoc] for detailed online docs.

Also note that the following packages, which were originally located in this repository,
have been moved to standalone repositories:

1. `base62` is a compact and high performance implementation of base62 algorithm for Golang.
   It has been moved to https://github.com/jxskiss/base62.

1. `extjson` extends the JSON syntax with extra features, itt helps in many scenes, such as:
   working with data generated by python programs, configuration with comments,
   managing many json files with duplicate contents, data driven testing by JSON files, etc.
   It has been moved to https://github.com/jxskiss/extjson.

1. `mcli` is a minimal but very powerful and magic cli library for Go.
   It has been moved to https://github.com/jxskiss/mcli.
