package monkey

import (
	"fmt"
	"reflect"
	"sync/atomic"

	"github.com/jxskiss/gopkg/v2/unsafe/forceexport"
	"github.com/jxskiss/gopkg/v2/unsafe/monkey/internal"
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
	targetMap  = make(map[any]*Patch)
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

	funcInfo *funcInfo
	varInfo  *varInfo
}

type funcInfo struct {
	proxy reflect.Value
	patch *internal.Patch
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

// Delete resets target to the state before applying the patch
// and destroys the patch.
// After calling Delete, the patch MUST NOT be used anymore.
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
func PatchFunc(target, repl any) *Patch {
	assertSameFuncType(target, repl)
	targetVal := reflect.ValueOf(target)
	replVal := reflect.ValueOf(repl)
	return patchFunc(targetVal, replVal)
}

// PatchMethod replaces a target's method with replacement.
// Replacement should expect the receiver (of type target) as the first argument.
// If the method cannot be found or the replacement type does not match, it panics.
func PatchMethod(target any, method string, repl any) *Patch {
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
func PatchByName(name string, repl any) *Patch {
	assertFunc(repl, "repl")
	targetPtr := forceexport.FindFuncWithName(name)
	targetVal := reflect.New(reflect.TypeOf(repl))
	forceexport.CreateFuncForCodePtr(targetVal.Interface(), targetPtr)
	return patchFunc(targetVal.Elem(), reflect.ValueOf(repl))
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

	proxy := reflect.New(target.Type())
	patch := internal.NewPatch(target, repl, proxy, false)
	p.funcInfo = &funcInfo{
		proxy: proxy,
		patch: patch,
	}

	targetMap[targetPtr] = p
	patchMap[p.id] = p
	return p
}

func (p *Patch) overrideFunc(repl reflect.Value) *Patch {
	child := &Patch{
		id:      newPatchID(),
		parent:  p,
		patched: false,
		target:  p.target,
		repl:    repl,
	}
	proxy := reflect.New(p.target.Type())
	patch := internal.NewPatch(p.target, repl, proxy, false)
	child.funcInfo = &funcInfo{
		proxy: proxy,
		patch: patch,
	}
	targetPtr := p.target.Pointer()
	targetMap[targetPtr] = child
	patchMap[child.id] = child
	return child
}

func (p *Patch) applyFunc() {
	if p.patched {
		p.funcInfo.patch.Patch()
	} else {
		p.funcInfo.patch.Unpatch(false)
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
	p.funcInfo.patch.Delete()
	if p.parent == nil {
		delete(targetMap, targetPtr)
	} else {
		targetMap[targetPtr] = p.parent
	}

	p.funcInfo = nil
	delete(patchMap, p.id)
}

// -------- patch variable -------- //

// PatchVar replaces target's value with replacement.
// If type of target and repl does not match, it panics.
func PatchVar(targetAddr, repl any) *Patch {
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
