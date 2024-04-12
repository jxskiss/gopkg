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

	// New goroutine will be created if len(queuedTasks) >= ScaleThreshold,
	// it defaults to 1, which means always start a new adhoc worker before
	// reaching the limit of total adhoc worker number.
	ScaleThreshold int

	// PermanentWorkerNum specifies the number of permanent workers to spawn
	// when creating a Pool, it defaults to 0 (no permanent worker).
	// Note that a permanent worker's goroutine stack is reused,
	// the memory won't be freed in the entire program lifetime.
	//
	// Generally you may want to set this to zero for common workloads,
	// tweak it for special workloads which benefits from reusing goroutine stacks.
	PermanentWorkerNum int

	// AdhocWorkerLimit specifies the initial limit of adhoc workers,
	// 0 or negative value means no limit.
	//
	// The limit of adhoc worker number can be changed by calling
	// SetAdhocWorkerLimit.
	AdhocWorkerLimit int

	// PanicHandler specifies a handler when panic occurs.
	// By default, a panic message with stack information is logged.
	PanicHandler func(context.Context, any)
}

// NewConfig creates a default Config.
func NewConfig() *Config {
	c := &Config{
		ScaleThreshold: 1,
	}
	c.checkAndSetDefaults()
	return c
}

func (c *Config) checkAndSetDefaults() {
	if c.PanicHandler == nil {
		c.PanicHandler = defaultPanicHandler
	}
}
