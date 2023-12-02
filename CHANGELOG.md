# Changelog

Notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

TBD.

## [2.11.0] - 2023-12-02

* Feat: [ezhttp] allow merge request data (#57)
* Feat: refactor [exp/mselect] and promote to be a stable package (#58, #60)
* Feat: refactor [exp/singleflight] and promote to be a stable package (#59)

## [2.10.0] - 2023-11-20

* Fix: [perf/bbp] drop cgo arena implementation, which doesn't work on Windows (#54)
* Feat: [utils/timeutil] new time related utilities
* Feat: [easy] new utility functions
* Change: drop support for Go 1.18, requires 1.19+
* Change: [utils/vdutil] update and simplify validating utilities (#56)
* Change: [perf/json] add new api `MarshalFastest`, and remove old api `MarshalNoMapOrdering`
  which is considered a bad design and there should be nobody using it

## [2.9.0] - 2023-08-11

* Feat: support go1.21
* Feat: new package [collection/heapx] with easy heap data structures
* Change: mark package [unsafe/monkey] as DEPRECATED

## [2.8.1] - 2023-05-20

* Fix: [perf/json] change to jsoniter as default JSON impl

## [2.8.0] - 2023-04-22

* Feat: [unsafe/monkey] support arm64 with bugfix and more tests (#48)
* Change: [perf/json] force fallback to jsoniter when bytedance/sonic is using compat implementation

## [2.7.5] - 2023-04-12

* Fix: [perf/json] build failure on arm64 (MacBook M1)

## [2.7.4] - 2023-04-10

* Feat: [utils/vdutil] new utility package helping data validation (#43)
* Feat: [zlog] utils to work with `context.Context` (#44)
* Feat: new util functions `easy.CreateNonExistingFolder` and `easy.WriteFile`

## [2.7.3] - 2023-03-21

* Feat: [utils/ptr] support converting integer to string pointer

## [2.7.1] - 2023-03-20

Minor improvements.

## [2.7.0] - 2023-03-06

* Feat: [perf/bbp] use slice to manage arena memory chunks which gives better performance
* Feat: [zlog] add logr support (#41)
* Feat: [zlog] tweak log format to be more friendly with the console format
* Fix: [perf/gopool] fix potential starving of task worker
* Change: Remove deprecated util functions `easy.All` and `easy.Any`

## [2.6.1] - 2023-03-03

* Feat: update CI workflow to support Go 1.20 (#37)
* Feat: update bytedance/sonic to v1.8.3
* Feat: [perf/bbp] new arena feature for perf/bbp (#40)
* Fix: code-style issues reported by golangci-lint (#37)
* Change: simplify perf/bbp implementation, notably that the signature of bbp.Grow and bbp.Buffer.Grow, bbp.Buffer.Append have been changed (#38)

## [2.5.1] - 2023-02-22

* Feat: new utility function `easy.InSlice`
* Feat: upgrade sonic to v1.8.1 to support Go 1.20 and better fallback strategy

## [2.5.0] - 2023-02-20

* Refactor: replace `interface{}` to `any` for cleaner code
* Feat: new utility function `easy.MergeMapsToPtr`
* Feat: [easy/ezhttp] allow to specify custom function to dump request/response
* Feat: fix `exp/perf/gopool` issues and move to `perf/gopool`
* Change: use value fields for `zlog.FileLogConfig`

## [2.4.0] - 2023-02-07

* Feat: [zlog] optimize std log redirection (#31)
* Feat: fork `github.com/bytedance/gopkg/util/gopool` with opinionated changes (#32)
* Feat: optimize performance of utility `easy.Unique`
* Feat: new utility function `easy.SumFloat`
* Fix: [perf/lru] `walbuf` does not reuse hashset memory
* Change: [perf/bbp] remove `ObjectPool`

## [2.3.6] - 2023-01-30

* Feat: [zlog] make param ctx being optional to `zlog.B`
* Feat: [zlog] export `zlog.Level`'s method `ToZapLevel`
* Feat: new utility function `easy.SplitMap` to split a large map into batches

## [2.3.4] - 20223-01-19

* Feat: [confr] supports custom logging function
* Feat: [confr] supports parsing `time.Duration` value

## [2.3.3] - 2023-01-10

* Fix: [perf/json] `HumanFriendly` indentation

## [2.3.2] - 2023-01-08

* Feat: http utility to reader response
* Feat: [perf/json] new config `HumanFriendly` which generates data that is more friendly
  for human reading
* Feat: new utility function `easy.LastIndex` to index value in a slice from last
* Change: rename `easy.ToHashSet` to `easy.ToBoolMap`

## [2.3.1] - 2023-01-07

* Feat: new util function `utils/ptr.NotZero`

## [2.3.0] - 2023-01-07

* Feat: new package `exp/mselect` to dynamically waiting many channels
* Feat: `easy.JSON` supports encoding map with `interface{}` as key type
* Feat: `perf/json` add new config option to use jsoniter as underlying implementation

## [2.2.0] - 2023-01-02

* Feat: rename package `easy/gemap` to `easy/ezmap`
* Feat: new `GetTyped` api for `easy/ezmap`
* Feat: ptr utils accept param of generic types
* Feat: change `unsafejson` implementation to bytedance/sonic (#29)

## [2.1.1] - 2022-12-10

* Feat: sync zap logfmt encoder from upstream
* Fix: bbp.indexPut get incorrect pool index

## [2.1.0] - 2022-12-10

Features:

* Feat: add support for Go 1.20
* Feat: update dependencies to latest
* Feat: update CI workflow to test under all linux/mac/windows platforms
* Feat: new api `easy.Glob` supporting double-star pattern
* Feat: new api `easy.LazyJSON` helps to avoid unnecessary marshaling in some use case
* Feat: new api `fastrand.Fastrand64`, `fastrand.ExpFloat64`, `fastrand.NormFloat64` and `fastrand.NewZipf`
* Feat: new api `zlog.Println`, `zlog.TRACE1` and `zlog.TRACESkip1`
* Feat: zlog support per logger output files
* Feat: optimize spinlock implementation
* Feat: make `errcode.Code` being compatible with `errors.Is`

Bugfix:

* Fix: test failures and bugs
* Fix: some issues reported by `golangci-lint`

Minor incompatible changes:

* Change return value of `monkey.Mock` to be exported type `Mocker`
* Remove deprecated type `json.InvalidUTF8Error`, `json.UnmarshalFieldError`,
  which are deprecated since Go 1.2
* Remove api `monkey.Patch.Origin`, which does not work correctly

## [2.0.0] - 2022-09-10

This release is a rewritten of the old v1 code, it features the generics feature
available from Go 1.18+.
