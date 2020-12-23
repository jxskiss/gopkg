package forceexport

import (
	"fmt"
	"reflect"
	"runtime"
	"unsafe"
)

// GetFunc gets the function defined by the given fully-qualified name.
// The outFuncPtr parameter should be a pointer to a function with the
// appropriate type (e.g. the address of a local variable), and is set to
// a new function value that calls the specified function.
// If the function does not exist, or is inlined, or inactive (won't be
// compiled into the binary), it panics.
func GetFunc(outFuncPtr interface{}, name string) {
	codePtr := FindFuncWithName(name)
	CreateFuncForCodePtr(outFuncPtr, codePtr)
}

// Func is a convenience struct for modifying the underlying code pointer
// of a function value. The actual struct has other values, but always
// starts with a code pointer.
type Func struct {
	codePtr uintptr
}

// CreateFuncForCodePtr accepts a code pointer and creates a function
// value that uses the pointer. The outFuncPtr argument should be a pointer
// to a function of the proper type (e.g. the address of a local variable),
// and will be set to the result function value.
func CreateFuncForCodePtr(outFuncPtr interface{}, codePtr uintptr) {
	outFuncVal := reflect.ValueOf(outFuncPtr).Elem()
	// Use reflect.MakeFunc to create a well-formed function value that's
	// guaranteed to be of the right type and guaranteed to be on the heap
	// (so that we can modify it).
	newFuncVal := reflect.MakeFunc(outFuncVal.Type(), nil)

	// Use reflection on the reflect.Value (yep!) to grab the underlying
	// function value pointer. Trying to call newFuncVal.Pointer() wouldn't
	// work because it gives the code pointer rather than the function value
	// pointer. The function value is a struct that starts with its code
	// pointer, so we can swap out the code pointer with our desired value.
	funcValuePtr := reflect.ValueOf(newFuncVal).FieldByName("ptr").Pointer()
	funcPtr := (*Func)(unsafe.Pointer(funcValuePtr))
	funcPtr.codePtr = codePtr

	outFuncVal.Set(newFuncVal)
}

// FindFuncWithName searches through the moduledata table created by the
// linker and returns the function's code pointer.
// If the function does not exist, or is inlined, or inactive (won't be
// compiled into the binary), it panics.
func FindFuncWithName(name string) uintptr {
	for _, moduleData := range activeModules() {
		for _, ftab := range moduleData.ftab {
			f := (*runtime.Func)(unsafe.Pointer(&moduleData.pclntable[ftab.funcoff]))
			if getName(f) == name {
				return f.Entry()
			}
		}
	}
	panic(fmt.Sprintf("forceexport: cannot find function %s, maybe inlined or inactive", name))
}

func getName(f *runtime.Func) string {
	defer func() {
		recover()
	}()
	return f.Name()
}

//go:linkname activeModules runtime.activeModules
func activeModules() []*moduledata

// Since the runtime.moduledata related data structures are not exported,
// we copy them below (and need to be consistent with the runtime).
type moduledata struct {
	pclntable []byte
	ftab      []functab

	// For usage in this package, we just need to ensure that
	// pclntable and ftab are consistent with runtime.moduledata.
	// ...
}

type functab struct {
	entry   uintptr
	funcoff uintptr
}

func init() {
	rtmdtype := GetType("runtime.moduledata")
	thismdtype := reflect.TypeOf(moduledata{})
	assertOffset(rtmdtype, thismdtype, "pclntable", "forceexport: moduledata.pclntable not match")
	assertOffset(rtmdtype, thismdtype, "ftab", "forceexport: moduledata.ftab not match")
}

func assertOffset(t1, t2 reflect.Type, fieldname string, msg string) {
	f1, ok1 := t1.FieldByName(fieldname)
	f2, ok2 := t2.FieldByName(fieldname)
	if !ok1 || !ok2 || f1.Offset != f2.Offset {
		panic(msg)
	}
}
