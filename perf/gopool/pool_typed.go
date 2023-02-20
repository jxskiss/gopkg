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
)

// TypedPool is a task-specific pool.
// A TypedPool is like pool, but it executes a handler to process values
// of a specific type.
// Compared to Pool, it helps to reduce unnecessary memory allocation of
// closures when submitting tasks.
type TypedPool[T any] struct {
	internalPool
}

// NewTypedPool creates a new task-specific pool with given handler and config.
func NewTypedPool[T any](config *Config, handler func(context.Context, T)) *TypedPool[T] {
	runner := newTypedTaskRunner(handler)
	p := &TypedPool[T]{}
	p.init(config, runner)
	return p
}

// Go submits a task to the pool.
func (p *TypedPool[T]) Go(arg T) {
	p.submit(context.Background(), arg)
}

// CtxGo submits a task to the pool, it's preferred over Go.
func (p *TypedPool[T]) CtxGo(ctx context.Context, arg T) {
	p.submit(ctx, arg)
}
