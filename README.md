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

1. `collection/heapx` contains a ready-to-use heap implementation based on `container/heap`,
   and a generic priority queue based on heap data structure.

1. `collection/set` provides generic set data structures.

1. `confr` provides a simple yet powerful configuration loader.

1. `easy` contains many handy utilities as a complementary to the standard library.

1. `easy/ezdbg` provides easy to use utilities which helps to do quick logging in development.

1. `easy/ezhttp` contains some utilities to work with HTTP requests.

1. `easy/ezmap` contains some utilities to manipulate map data structure.

1. `easy/yamlx` extends the YAML unmarshaler with extra features, such as
   "reading environment variables" and "file including",
   "reference using gjson JSON path expression", "reference using named variables",
   "function calling", etc.

1. `encrypt/crypto` contains some encryption utilities.

1. `exp/kvutil` contains experimental utilities to work with key-value cache.

1. `infra/acache` contains an async cache which can be used to fetch and update the
   latest data periodically and supports expiring a key if it's unused for a period.

1. `infra/errcode` provides error code registry to manage error codes and messages.
   Standardized error codes greatly helps in large scale microservices.

1. `infra/logid` provides log ID generators to help service observability, such as
   tracing, metrics, etc.

1. `perf/bbp` provides efficient byte buffer pools with anti-memory-waste protection.

1. `perf/gopool` is a fork of `github.com/bytedance/gopkg/util/gopool`
   with opinionated changes.
   It's a high-performance goroutine pool which aims to reuse goroutines
   and limit the number of goroutines.

1. `perf/json` provides a drop-in replacement of `encoding/json` and extended features.
   By default, it uses [jsoniter] in underlying, which has better performance than `encoding/json`.
   `amd64` users may use `json.Config` to switch to [bytedance/sonic] for best performance
   in the cost of a little less maturity.
   Check [README.md](./perf/json/README.md) for detailed introduction.

1. `perf/lru` is a high performance implementation of the LRU cache, it features pre-allocation,
   item expiration, friendly and type-safe APIs for commonly used key types.
   It also provides a sharded version for heavy lock contention use-case.

1. `unsafe/forceexport` force exports private types and functions of other packages. As you might
   expect, this package is unsafe and fragile and shouldn't be used in production, it is
   mainly for testing purpose.

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

1. `utils/timeutil` provides utilities to process time related things.

1. `utils/vdutil` helps to do data validation.

1. `zlog` provides opinionated high-level logging facilities based on go.uber.org/zap.

See [GoDoc][godoc] for detailed online docs.

[bytedance/sonic]: https://github.com/bytedance/sonic
[jsoniter]: https://github.com/json-iterator/go
[zap]: https://github.com/uber-go/zap

Also note that the following packages, which were originally located in this repository,
have been moved to standalone repositories:

1. `base62` is a compact and high performance implementation of base62 algorithm for Golang.
   It has been moved to https://github.com/jxskiss/base62.

1. `mcli` is a minimal but very powerful and magic cli library for Go.
   It has been moved to https://github.com/jxskiss/mcli.
