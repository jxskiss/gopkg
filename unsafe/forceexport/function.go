package forceexport

import (
	"fmt"
	"reflect"
	"runtime"
	"unsafe"

	"github.com/jxskiss/gopkg/v2/internal/linkname"
)

// GetFunc gets the function defined by the given fully-qualified name.
// The outFuncPtr parameter should be a pointer to a function with the
// appropriate type (e.g. the address of a local variable), and is set to
// a new function value that calls the specified function.
// If the function does not exist, or is inlined, or inactive (haven't been
// compiled into the binary), it panics.
func GetFunc(outFuncPtr any, name string) {
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
func CreateFuncForCodePtr(outFuncPtr any, codePtr uintptr) {
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

	funcPtr := (*Func)(unsafe.Pointer(
		reflect.ValueOf(newFuncVal).FieldByName("ptr").Pointer(),
	))
	funcPtr.codePtr = codePtr

	outFuncVal.Set(newFuncVal)
}

// FindFuncWithName searches through the moduledata table created by the
// linker and returns the function's code pointer.
// If the function does not exist, or is inlined, or inactive (haven't been
// compiled into the binary), it panics.
func FindFuncWithName(name string) uintptr {
	modulesSlice := linkname.Runtime_activeModules()
	for _, md := range modulesSlice {
		pclntable := md.Field_pclntable()
		for _, ftab := range md.Field_ftab() {
			f := (*runtime.Func)(unsafe.Pointer(&pclntable[ftab.Field_funcoff()]))
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
