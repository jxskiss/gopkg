package monkey

import (
	"fmt"
	"github.com/jxskiss/gopkg/forceexport"
	"github.com/jxskiss/gopkg/internal/linkname"
	"reflect"
)

var patchTable = make(map[uintptr]*PatchGuard)

// PatchGuard holds the patch status of a patch target and it's replacement.
type PatchGuard struct {
	patched     bool
	target      reflect.Value
	replacement reflect.Value
	origBytes   []byte
	replBytes   []byte
}

// Unpatch removes the monkey patch of the target.
func (p *PatchGuard) Unpatch() {
	unpatchValue(p.target.Pointer())
}

// Restore re-apply the monkey patch to the target if removed.
// If the patch has already been applied, it's a no-op.
func (p *PatchGuard) Restore() {
	patchValue(p.target, p.replacement)
}

// Patch replaces a function with replacement.
// If target or replacement is not a function their types do not match, it panics.
func Patch(target, replacement interface{}) *PatchGuard {
	return patchValue(reflect.ValueOf(target), reflect.ValueOf(replacement))
}

// PatchMethod replaces an target's methodName method with replacement.
// Replacement should expect the receiver (of type target) as the first argument.
// If the method cannot be found or the replacement type does not match, it panics.
func PatchMethod(target interface{}, methodName string, replacement interface{}) *PatchGuard {
	targetTyp := reflect.TypeOf(target)
	method, ok := targetTyp.MethodByName(methodName)
	if !ok {
		panic(fmt.Sprintf("monkey: unknown method %s.%s", targetTyp.Name(), methodName))
	}
	return patchValue(method.Func, reflect.ValueOf(replacement))
}

// PatchByName replaces a function with replacement by it's name.
// TargetName should be the fully-qualified name of the target function or method.
// If the target cannot be found or the replacement type does not match, it panics.
func PatchByName(targetName string, replacement interface{}) *PatchGuard {
	targetPtr := forceexport.FindFuncWithName(targetName)
	targetVal := reflect.New(reflect.TypeOf(replacement))
	forceexport.CreateFuncForCodePtr(targetVal.Interface(), targetPtr)
	return patchValue(targetVal.Elem(), reflect.ValueOf(replacement))
}

// Unpatch removes any monkey patch applied to the target.
// It returns whether the target was patched in the first place.
func Unpatch(target interface{}) bool {
	return unpatchValue(reflect.ValueOf(target).Pointer())
}

// UnpatchMethod removes any monkey patch applied to the methodName method of the target.
// It returns whether the method was patched in the first place.
func UnpatchMethod(target interface{}, methodName string) bool {
	targetTyp := reflect.TypeOf(target)
	method, ok := targetTyp.MethodByName(methodName)
	if !ok {
		panic(fmt.Sprintf("monkey: unknown method %s.%s", targetTyp.Name(), methodName))
	}
	return unpatchValue(method.Func.Pointer())
}

// UnpatchByName removes any monkey patch applied to the target by it's name.
// TargetName should be the fully-qualified name of the target function or method.
func UnpatchByName(targetName string) bool {
	targetPtr := forceexport.FindFuncWithName(targetName)
	return unpatchValue(targetPtr)
}

// UnpatchAll removes all applied monkey patches.
func UnpatchAll() {
	linkname.Runtime_stopTheWorld()
	for target, patch := range patchTable {
		if !patch.patched {
			continue
		}
		replaceCode(target, patch.origBytes)
		patch.patched = false
	}
	linkname.Runtime_startTheWorld()
}

func patchValue(target, replacement reflect.Value) *PatchGuard {
	if target.Kind() != reflect.Func {
		panic("monkey: target is not a function")
	}
	if replacement.Kind() != reflect.Func {
		panic("monkey: replacement is not a function")
	}
	if target.Type() != replacement.Type() {
		panic("monkey: target and replacement have different types")
	}

	linkname.Runtime_stopTheWorld()
	targetPtr := target.Pointer()
	patch, ok := patchTable[targetPtr]
	if !ok {
		patch = &PatchGuard{}
		patchTable[targetPtr] = patch
	}
	if !patch.replacement.IsValid() || patch.replacement != replacement {
		replPtr := uintptr(getPtr(replacement))
		replBytes := buildJmpDirective(replPtr)
		patch.replacement = replacement
		patch.replBytes = replBytes
	}
	if len(patch.origBytes) == 0 {
		patchSize := len(patch.replBytes)
		patch.target = target
		patch.origBytes = copy_(getCode(targetPtr, patchSize))
	}
	replaceCode(targetPtr, patch.replBytes)
	patch.patched = true
	linkname.Runtime_startTheWorld()
	return patch
}

func unpatchValue(target uintptr) bool {
	patch, ok := patchTable[target]
	if !ok || !patch.patched {
		return false
	}

	linkname.Runtime_stopTheWorld()
	replaceCode(target, patch.origBytes)
	patch.patched = false
	linkname.Runtime_startTheWorld()
	return true
}

// this function is super unsafe
// It copies a slice to a raw memory location, disabling all memory protection before doing so.
func replaceCode(target uintptr, code []byte) {
	_replace_code(target, code)
}
