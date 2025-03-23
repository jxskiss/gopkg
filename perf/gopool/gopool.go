// Copyright 2021 ByteDance Inc.
// Copyright 2023 Shawn Wang <jxskiss@126.com>.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package gopool is a high-performance goroutine pool which aims to reuse goroutines
// and limit the number of goroutines.
package gopool

import (
	"context"
)

// defaultPool is the default unbounded pool.
var defaultPool *Pool

func init() {
	config := &Config{Name: "gopool.defaultPool"}
	defaultPool = NewPool(config)
}

// Go is an alternative to the go keyword, which is able to recover panic,
// reuse goroutine stack, limit goroutine numbers, etc.
//
// See package doc for detailed introduction.
func Go(f func()) {
	defaultPool.CtxGo(context.Background(), f)
}

// CtxGo is preferred over Go.
func CtxGo(ctx context.Context, f func()) {
	defaultPool.CtxGo(ctx, f)
}
