/*
 * Copyright 2022 ByteDance Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package common

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSysAlloc(t *testing.T) {
	assert.NotPanics(t, func() {
		data := AllocatePage()

		// try write first and last byte of data
		data[0] = 0
		data[len(data)-1] = 0

		// try get mem stat
		m := runtime.MemStats{}
		runtime.ReadMemStats(&m)

		ReleasePage(data)
	})
}
