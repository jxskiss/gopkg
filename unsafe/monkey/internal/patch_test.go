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

package internal

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jxskiss/gopkg/v2/internal/testutil"
)

//go:noinline
func Target() int {
	return 0
}

func Hook() int {
	return 2
}

func UnsafeTarget() {}

func TestPatchFunc(t *testing.T) {
	if !testutil.IsDisableInlining() {
		t.Skip("skip: inlining not disabled")
	}

	t.Run("normal", func(t *testing.T) {
		var proxy func() int
		patch := patchFunc(Target, Hook, &proxy, false)
		assert.Equal(t, 2, Target())
		assert.Equal(t, 0, proxy())

		patch.Unpatch(true)
		assert.Equal(t, 0, Target())
	})

	t.Run("anonymous hook", func(t *testing.T) {
		var proxy func() int
		patch := patchFunc(Target, func() int { return 2 }, &proxy, false)
		assert.Equal(t, 2, Target())
		assert.Equal(t, 0, proxy())

		patch.Unpatch(true)
		assert.Equal(t, 0, Target())
	})

	t.Run("closure hook", func(t *testing.T) {
		var proxy func() int
		hookBuilder := func(x int) func() int {
			return func() int { return x }
		}
		patch := patchFunc(Target, hookBuilder(2), &proxy, false)
		assert.Equal(t, 2, Target())
		assert.Equal(t, 0, proxy())

		patch.Unpatch(true)
		assert.Equal(t, 0, Target())
	})

	t.Run("reflect hook", func(t *testing.T) {
		var proxy func() int
		hookVal := reflect.MakeFunc(reflect.TypeOf(Hook), func(args []reflect.Value) (results []reflect.Value) { return []reflect.Value{reflect.ValueOf(2)} })
		patch := patchFunc(Target, hookVal.Interface(), &proxy, false)
		assert.Equal(t, 2, Target())
		assert.Equal(t, 0, proxy())

		patch.Unpatch(true)
		assert.Equal(t, 0, Target())
	})

	t.Run("unsafe", func(t *testing.T) {
		var proxy func()
		patch := patchFunc(UnsafeTarget, func() { panic("good") }, &proxy, true)
		assert.PanicsWithValue(t, "good", func() { UnsafeTarget() })
		assert.NotPanics(t, func() { proxy() })

		patch.Unpatch(true)
		assert.NotPanics(t, func() { UnsafeTarget() })
	})
}
