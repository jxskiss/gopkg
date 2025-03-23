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
	"log"
	"time"

	"github.com/jxskiss/gopkg/v2/internal"
)

func defaultPanicHandler(_ context.Context, exc any) {
	location, frames := internal.IdentifyPanic(1)
	log.Printf("[ERROR] gopool: catch panic: %v\nlocation: %v\n%s\n", exc, location, internal.FormatFrames(frames))
}

// Config is used to config a Pool instance.
type Config struct {

	// Name optionally specifies the name of a pool instance.
	Name string

	// PanicHandler specifies a handler when panic occurs.
	// By default, a panic message with stack information is logged.
	PanicHandler func(context.Context, any)

	bounded *bool

	// Bounded goroutine pool options.
	scaleThreshold     int
	permanentWorkerNum int
	adhocWorkerLimit   int

	// Unbounded goroutine pool options.
	maxIdle int32
	maxAge  time.Duration
}

// NewConfig creates a default Config.
func NewConfig() *Config {
	return &Config{}
}

// SetBounded configures a pool to start limit number of goroutines.
//
// New goroutine will be created if len(queuedTasks) >= scaleThreshold,
// it defaults to 1, which means always start a new adhoc worker before
// reaching the adhoc worker limit.
//
// permanentWorkerNum specifies the number of permanent workers to spawn
// when creating a pool, it defaults to 0 (no permanent worker).
// Note that a permanent worker goroutine's stack is reused,
// the memory won't be freed in the entire program lifetime.
// Generally you may want to set this to zero for common workloads,
// tweak it for special workloads which benefits from reusing goroutine stacks.
//
// adhocWorkerLimit specifies the initial limit of adhoc workers,
// 0 or negative value means no limit.
// The limit of adhoc worker number can be changed by calling
// SetAdhocWorkerLimit.
func (c *Config) SetBounded(scaleThreshold, permanentWorkerNum, adhocWorkerLimit int) *Config {
	c.bounded = toptr(true)
	c.scaleThreshold = scaleThreshold
	c.permanentWorkerNum = permanentWorkerNum
	c.adhocWorkerLimit = adhocWorkerLimit
	return c
}

// SetUnbounded configures a pool to start unlimited number of goroutines when needed.
//
// maxIdle is the max idle workers keeping in pool for waiting tasks.
// goroutines exceed maxIdle exit immediately after finish a task.
// It defaults to 1000.
//
// workerMaxAge specifies the max age of a worker in pool.
// It defaults to 1 minute.
func (c *Config) SetUnbounded(maxIdle int, workerMaxAge time.Duration) *Config {
	c.bounded = toptr(false)
	c.maxIdle = int32(maxIdle)
	c.maxAge = workerMaxAge
	return c
}

func (c *Config) checkAndSetBoundedDefaults() {
	if c.scaleThreshold <= 0 {
		c.scaleThreshold = 1
	}
	if c.PanicHandler == nil {
		c.PanicHandler = defaultPanicHandler
	}
}

func (c *Config) checkAndSetUnboundedDefaults() {
	if c.maxIdle <= 0 {
		c.maxIdle = 1000
	}
	if c.maxAge <= 0 {
		c.maxAge = time.Minute
	}
	if c.PanicHandler == nil {
		c.PanicHandler = defaultPanicHandler
	}
}

func (c *Config) isUnbounded() bool {
	return c.bounded == nil || !*c.bounded
}

func toptr[T any](v T) *T { return &v }
