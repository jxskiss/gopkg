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

	"github.com/jxskiss/gopkg/v2/unsafe/monkey/internal/common"
	"github.com/jxskiss/gopkg/v2/unsafe/monkey/internal/fn"
	"github.com/jxskiss/gopkg/v2/unsafe/monkey/internal/inst"
	"github.com/jxskiss/gopkg/v2/unsafe/monkey/internal/mem"
	"github.com/jxskiss/gopkg/v2/unsafe/monkey/internal/tool"
)

// Patch is a context that holds the address and original codes of the patched function.
type Patch struct {
	size int
	code []byte
	repl []byte
	base uintptr
}

// Delete releases memory to the operating system.
func (p *Patch) Delete() {
	common.ReleasePage(p.code)
}

// Patch replaces target function code with the patch code.
func (p *Patch) Patch() {
	mem.WriteWithSTW(p.base, p.repl[:p.size])
}

// Unpatch restores the patched function to the original function.
func (p *Patch) Unpatch(delete bool) {
	mem.WriteWithSTW(p.base, p.code[:p.size])
	if delete {
		p.Delete()
	}
}

// NewPatch replace the target function with a hook function, and stores the target function in the proxy function
// for future restore. Target and hook are values of function. Proxy is a value of proxy function pointer.
func NewPatch(target, hook, proxy reflect.Value, unsafe bool) *Patch {
	tool.Assert(hook.Kind() == reflect.Func, "'%s' is not a function", hook.Kind())
	tool.Assert(proxy.Kind() == reflect.Ptr, "'%v' is not a function pointer", proxy.Kind())
	tool.Assert(hook.Type() == target.Type(), "'%v' and '%s' mismatch", hook.Type(), target.Type())
	tool.Assert(proxy.Elem().Type() == target.Type(), "'*%v' and '%s' mismatch", proxy.Elem().Type(), target.Type())

	targetAddr := target.Pointer()
	// The first few bytes of the target function code
	const bufSize = 64
	targetCodeBuf := common.BytesOf(targetAddr, bufSize)
	// construct the branch instruction, i.e. jump to the hook function
	hookCode := inst.BranchInto(common.PtrAt(hook))
	// search the cutting point of the target code, i.e. the minimum length of full instructions that is longer than the hookCode
	cuttingIdx := inst.Disassemble(targetCodeBuf, len(hookCode), !unsafe)

	// construct the proxy code
	proxyCode := common.AllocatePage()
	// save the original code before the cutting point
	copy(proxyCode, targetCodeBuf[:cuttingIdx])
	// construct the branch instruction, i.e. jump to the cutting point
	copy(proxyCode[cuttingIdx:], inst.BranchTo(targetAddr+uintptr(cuttingIdx)))
	// inject the proxy code to the proxy function
	fn.InjectInto(proxy, proxyCode)

	tool.DebugPrintf("NewPatch: hook code len(%v), cuttingIdx(%v)\n", len(hookCode), cuttingIdx)

	return &Patch{base: targetAddr, code: proxyCode, repl: hookCode, size: cuttingIdx}
}

func patchFunc(fn, hook, proxy interface{}, unsafe bool) *Patch {
	vv := reflect.ValueOf(fn)
	tool.Assert(vv.Kind() == reflect.Func, "'%v' is not a function", fn)
	patch := NewPatch(vv, reflect.ValueOf(hook), reflect.ValueOf(proxy), unsafe)
	patch.Patch()
	return patch
}
