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
	"runtime/debug"

	"github.com/jxskiss/gopkg/v2/internal"
	"github.com/jxskiss/gopkg/v2/zlog"
)

const (
	defaultScaleThreshold = 1
)

func defaultPanicHandler(_ context.Context, exc any) {
	loc := internal.IdentifyPanic(1)
	zlog.StdLogger.Errorf("gopool: catch panic: %v, location: %v\n%s\n", exc, loc, debug.Stack())
}

// Config is used to config a Pool instance.
type Config struct {

	// Name optionally specifies the name of a pool instance.
	Name string

	// New goroutine will be created if len(queued tasks) > ScaleThreshold,
	// it defaults to 1.
	ScaleThreshold int

	// PanicHandler specifies a handler when panic occurs.
	// By default, a panic message with stack information is logged.
	PanicHandler func(context.Context, any)

	// PermanentWorkerNum specifies the number of permanent workers to spawn
	// when creating a Pool, it defaults to 0 (no permanent worker).
	// Note that permanent workers' goroutine stack will be reused,
	// the memory won't be freed in the entire program lifetime.
	//
	// Generally you may want to set this to zero for common workloads,
	// tweak it for special workloads which benefits from reusing goroutine stacks.
	PermanentWorkerNum int

	// AdhocWorkerLimit specifies the initial limit of adhoc workers,
	// 0 or negative value means no limit.
	//
	// The limit of adhoc worker number can be changed by calling
	// Pool.SetAdhocWorkerLimit.
	AdhocWorkerLimit int
}

// NewConfig creates a default Config.
func NewConfig() *Config {
	c := &Config{}
	c.checkAndSetDefaults()
	return c
}

func (c *Config) checkAndSetDefaults() {
	if c.ScaleThreshold <= 0 {
		c.ScaleThreshold = defaultScaleThreshold
	}
	if c.PanicHandler == nil {
		c.PanicHandler = defaultPanicHandler
	}
}
