// Copyright 2025 CloudWeGo Authors
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

// Package gopool provides goroutine pool that helps to reuse goroutines
// for better performance.
package gopool

import (
	"context"
	"fmt"
	"time"

	"github.com/jxskiss/gopkg/v2/internal"
)

// Option configures the behavior of a GoPool.
type Option struct {
	// MaxIdleWorkers is the max number of idle workers keeping alive for waiting tasks.
	// Idle workers exit after WorkerMaxAge.
	MaxIdleWorkers int

	// WorkerMaxAge is the max age of a worker.
	WorkerMaxAge time.Duration

	// TaskChanBuffer is the size of task queue channel.
	// If the queue is full, we will fall back to use `go` directly without using pool.
	// Normally, the queue length should be small,
	// coz we create new workers to pick tasks if necessary.
	TaskChanBuffer int

	// PanicHandler specifies a handler when panic occurs.
	// By default, a panic message with stack information is logged.
	PanicHandler PanicHandler
}

type PanicHandler func(ctx context.Context, r any)

// DefaultOption returns the default option.
func DefaultOption() *Option {
	return &Option{
		MaxIdleWorkers: 1000,
		WorkerMaxAge:   time.Minute,
		TaskChanBuffer: 1000,
	}
}

var defaultPanicHandler = func(ctx context.Context, r any) {
	location, frames := internal.IdentifyPanic(1)
	err := fmt.Errorf("%v", r)
	msg := fmt.Sprintf("gopool: catch panic: %v\nlocation: %v\n%s\n", r, location, internal.FormatFrames(frames))
	internal.DefaultLoggerError(ctx, err, msg)
}

var defaultPool = New("gopool.defaultPool", nil)

// Go runs the given func in background
func Go(f func()) {
	defaultPool.CtxGo(context.Background(), f)
}

// CtxGo runs the given func in background, and it passes ctx to panic handler when happens.
func CtxGo(ctx context.Context, f func()) {
	defaultPool.CtxGo(ctx, f)
}

// SetDefaultPanicHandler sets a default panic handler.
//
// Check GoPool.SetPanicHandler for changing panic handler of an individual pool.
func SetDefaultPanicHandler(handler PanicHandler) {
	defaultPanicHandler = handler
}
