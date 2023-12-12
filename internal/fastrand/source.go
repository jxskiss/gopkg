// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fastrand

import "github.com/jxskiss/gopkg/v2/internal/linkname"

// globalSource is the source of random numbers for the top-level
// convenience functions.
var globalSource = &runtimeSource{}

// A Source is a source of uniformly-distributed
// pseudo-random uint64 values in the range [0, 1<<64).
//
// A Source is not safe for concurrent use by multiple goroutines.
type Source interface {
	Uint64() uint64
}

// runtimeSource is a Source that uses the runtime fastrand functions.
type runtimeSource struct{}

func (*runtimeSource) Uint64() uint64 {
	return linkname.Runtime_fastrand64()
}
