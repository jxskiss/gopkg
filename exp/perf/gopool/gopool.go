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

package gopool

import (
	"context"
	"fmt"
	"sync"
)

var defaultPool *Pool

var poolMap sync.Map

func init() {
	config := &Config{
		Name:             "gopool.defaultPool",
		AdhocWorkerLimit: 10000,
	}
	defaultPool = NewPool(config)
}

// Default returns the global default pool.
// The default pool does not enable permanent workers,
// the adhoc worker limit is configured to 10000.
// The package-level methods Go and CtxGo submit tasks to the default pool.
//
// Note that it's not recommended to change the worker limit
// of the default pool, which affects other code that use the default pool.
func Default() *Pool {
	return defaultPool
}

// Go is an alternative to the go keyword, which is able to recover panic,
// reuse goroutine stack, limit goroutine numbers, etc.
//
// See package doc for detailed introduction.
func Go(f func()) {
	CtxGo(context.Background(), f)
}

// CtxGo is preferred over Go.
func CtxGo(ctx context.Context, f func()) {
	defaultPool.CtxGo(ctx, f)
}

// Register registers a Pool to the global map,
// it returns error if the same name has already been registered.
// To register a pool, the pool should be configured with a
// non-empty name.
//
// Get can be used to get the registered pool by name.
func Register(p *Pool) error {
	_, loaded := poolMap.LoadOrStore(p.Name(), p)
	if loaded {
		return fmt.Errorf("gopool: %s already registered", p.Name())
	}
	return nil
}

// Get gets a registered Pool by name.
// It returns nil if specified pool is not registered.
func Get(name string) *Pool {
	p, ok := poolMap.Load(name)
	if !ok {
		return nil
	}
	return p.(*Pool)
}
