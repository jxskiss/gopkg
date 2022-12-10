# Changelog

Notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

TBD.

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
