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

// Pool manages a goroutine pool and tasks for better performance,
// it reuses goroutines and limits the number of goroutines.
type Pool = TypedPool[func()]

// NewPool creates a new pool with the config.
func NewPool(config *Config) *Pool {
	runner := funcTaskRunner
	p := &TypedPool[func()]{}
	p.init(config, runner)
	return p
}
