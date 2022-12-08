package monkey

import (
	"bytes"
	"fmt"
	"reflect"
	"sync/atomic"

	"github.com/jxskiss/gopkg/v2/internal/linkname"
	"github.com/jxskiss/gopkg/v2/unsafe/forceexport"
)

// AutoUnpatch encapsulates a function with a context, which automatically
// unpatch all patches applied within function f.
func AutoUnpatch(f func()) {
	defer popPatchStack()
	addPatchStack()
	f()
}

var (
	idCounter  int64
	patchMap   = make(map[int64]*Patch)
	targetMap  = make(map[interface{}]*Patch)
	patchStack [][]int64
)

func newPatchID() int64 {
	return atomic.AddInt64(&idCounter, 1)
}

func addPatchStack() {
	patchStack = append(patchStack, make([]int64, 0))
}

func popPatchStack() {
	if len(patchStack) == 0 {
		return
	}
	patchIDs := patchStack[len(patchStack)-1]
	for i := len(patchIDs) - 1; i >= 0; i-- {
		id := patchIDs[i]
		p := patchMap[id]
		p.Delete()
	}
	patchStack = patchStack[:len(patchStack)-1]
}

// Patch holds the patch data of a patch target and its replacement.
type Patch struct {
	id      int64
	parent  *Patch
	patched bool
	target  reflect.Value
	repl    reflect.Value

	*funcInfo
	*varInfo
}

type funcInfo struct {
	replCode  []byte
	origCode  []byte
	branchPos int
}

type varInfo struct {
	origVar reflect.Value
}

func (p *Patch) apply() {
	if p.funcInfo != nil {
		p.applyFunc()
	} else if p.varInfo != nil {
		p.applyVar()
	}
}

// Patch applies the patch.
func (p *Patch) Patch() {
	p.patched = true
	p.apply()
}

// Delete resets target to the state before applying the patch.
func (p *Patch) Delete() {
	if p.funcInfo != nil {
		p.deleteFunc()
	} else if p.varInfo != nil {
		p.deleteVar()
	}
}

// -------- patch function -------- //

// PatchFunc replaces a function with replacement.
// If target or replacement is not a function or their types do not match, it panics.
func PatchFunc(target, repl interface{}) *Patch {
	assertSameFuncType(target, repl)
	targetVal := reflect.ValueOf(target)
	replVal := reflect.ValueOf(repl)
	return patchFunc(targetVal, replVal)
}

// PatchMethod replaces a target's method with replacement.
// Replacement should expect the receiver (of type target) as the first argument.
// If the method cannot be found or the replacement type does not match, it panics.
func PatchMethod(target interface{}, method string, repl interface{}) *Patch {
	assertFunc(repl, "repl")
	targetTyp := reflect.TypeOf(target)
	targetMethod, ok := targetTyp.MethodByName(method)
	if !ok {
		panic(fmt.Sprintf("monkey: unknown method %s.%s", targetTyp.Name(), method))
	}
	return patchFunc(targetMethod.Func, reflect.ValueOf(repl))
}

// PatchByName replaces a function with replacement by it's name.
// TargetName should be the fully-qualified name of the target function or method.
// If the target cannot be found or the replacement type does not match, it panics.
func PatchByName(name string, repl interface{}) *Patch {
	assertFunc(repl, "repl")
	targetPtr := forceexport.FindFuncWithName(name)
	targetVal := reflect.New(reflect.TypeOf(repl))
	forceexport.CreateFuncForCodePtr(targetVal.Interface(), targetPtr)
	return patchFunc(targetVal.Elem(), reflect.ValueOf(repl))
}

// replaceCode copies a slice to a raw memory location,
// disabling all memory protection before doing so.
//
// As it sounds, this function is super unsafe.
func replaceCode(target uintptr, code []byte) {
	linkname.Runtime_stopTheWorld()
	_replace_code(target, code)
	linkname.Runtime_startTheWorld()
}

func patchFunc(target, repl reflect.Value) *Patch {
	assertSameFuncType(target.Interface(), repl.Interface())
	patch := newFuncPatch(target, repl)
	patch.Patch()

	if len(patchStack) > 0 {
		patchIDs := &patchStack[len(patchStack)-1]
		*patchIDs = append(*patchIDs, patch.id)
	}

	return patch
}

func newFuncPatch(target, repl reflect.Value) *Patch {
	targetPtr := target.Pointer()
	if old := targetMap[targetPtr]; old != nil {
		return old.overrideFunc(repl)
	}

	p := &Patch{
		id:      newPatchID(),
		parent:  nil,
		patched: false,
		target:  target,
		repl:    repl,
	}

	targetCode := getCode(targetPtr, 64)
	replCode := branchInto(uintptr(getPtr(repl)))
	branchPos := disassemble(targetCode, len(replCode))

	origCode := linkname.Runtime_sysAlloc(_PAGE_SIZE)
	origSize := copy(origCode, targetCode[:branchPos])
	origCode = origCode[:origSize]

	// protect the memory, avoid SIGBUS unexpected fault address error
	_replace_code(slicePtr(origCode), origCode)

	p.funcInfo = &funcInfo{
		replCode:  replCode,
		origCode:  origCode,
		branchPos: branchPos,
	}

	targetMap[targetPtr] = p
	patchMap[p.id] = p
	return p
}

func (p *Patch) overrideFunc(repl reflect.Value) *Patch {
	replCode := branchInto(uintptr(getPtr(repl)))
	child := &Patch{
		id:      newPatchID(),
		parent:  p,
		patched: false,
		target:  p.target,
		repl:    repl,
	}
	child.funcInfo = &funcInfo{
		replCode:  replCode,
		origCode:  p.origCode,
		branchPos: p.branchPos,
	}
	targetPtr := p.target.Pointer()
	targetMap[targetPtr] = child
	patchMap[child.id] = child
	return child
}

func (p *Patch) applyFunc() {
	targetPtr := p.target.Pointer()
	if p.patched {
		code := getCode(targetPtr, len(p.replCode))
		if !bytes.Equal(code, p.replCode) {
			replaceCode(targetPtr, p.replCode)
		}
	} else {
		code := getCode(targetPtr, p.branchPos)
		origCode := p.origCode[:p.branchPos]
		if !bytes.Equal(code, origCode) {
			replaceCode(targetPtr, origCode)
		}
	}
}

func (p *Patch) deleteFunc() {
	p.patched = false
	if p.parent != nil {
		p.parent.apply()
	} else {
		p.apply()
	}

	targetPtr := p.target.Pointer()
	if p.parent == nil {
		linkname.Runtime_sysFree(p.origCode)
		delete(targetMap, targetPtr)
	} else {
		targetMap[targetPtr] = p.parent
	}
	delete(patchMap, p.id)
}

// -------- patch variable -------- //

// PatchVar replaces target's value with replacement.
// If type of target and repl does not match, it panics.
func PatchVar(targetAddr, repl interface{}) *Patch {
	assertVarPtr(targetAddr)
	if repl == nil {
		repl = reflect.Zero(reflect.TypeOf(targetAddr).Elem())
	}
	targetAddrVal := reflect.ValueOf(targetAddr)
	replVal := reflect.ValueOf(repl)
	assertVarReplacement(targetAddrVal, replVal)

	patch := newVarPatch(targetAddrVal, replVal)
	patch.Patch()

	if len(patchStack) > 0 {
		patchIDs := &patchStack[len(patchStack)-1]
		*patchIDs = append(*patchIDs, patch.id)
	}

	return patch
}

func newVarPatch(targetAddr, repl reflect.Value) *Patch {
	if old := targetMap[targetAddr]; old != nil {
		return old.overrideVar(repl)
	}

	p := &Patch{
		id:      newPatchID(),
		parent:  nil,
		patched: false,
		target:  targetAddr,
		repl:    repl,
	}

	orig := reflect.New(targetAddr.Type().Elem()).Elem()
	orig.Set(targetAddr.Elem())
	p.varInfo = &varInfo{
		origVar: orig,
	}

	targetMap[targetAddr] = p
	patchMap[p.id] = p
	return p
}

func (p *Patch) overrideVar(repl reflect.Value) *Patch {
	child := &Patch{
		id:      newPatchID(),
		parent:  p,
		patched: false,
		target:  p.target,
		repl:    repl,
	}
	child.varInfo = &varInfo{
		origVar: p.varInfo.origVar,
	}
	targetMap[p.target] = child
	patchMap[child.id] = child
	return child
}

func (p *Patch) applyVar() {
	if p.patched {
		p.target.Elem().Set(p.repl)
	} else {
		p.target.Elem().Set(p.varInfo.origVar)
	}
}

func (p *Patch) deleteVar() {
	p.patched = false
	if p.parent != nil {
		p.parent.apply()
	} else {
		p.apply()
	}

	if p.parent == nil {
		delete(targetMap, p.target.Interface())
	} else {
		targetMap[p.target.Interface()] = p.parent
	}
	delete(patchMap, p.id)
}
