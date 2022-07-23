// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package singleflight provides utilities wrapping around
// package golang.org/x/sync/singleflight, which provides a
// duplicate function call suppression mechanism.
package singleflight

import "golang.org/x/sync/singleflight"

// Group is an alias name of golang.org/x/sync/singleflight.Group.
type Group = singleflight.Group

// Result is an alias name of golang.org/x/sync/singleflight.Result.
type Result = singleflight.Result
