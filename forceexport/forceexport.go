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
// a new function value that calls the specified function. If the specified
// function does not exist, outFuncPtr is not set and an error is returned.
func GetFunc(outFuncPtr interface{}, name string) error {
	codePtr, err := FindFuncWithName(name)
	if err != nil {
		return err
	}
	CreateFuncForCodePtr(outFuncPtr, codePtr)
	return nil
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
// linker and returns the function's code pointer. If the function was not
// found, it returns an error. Since the data structures here are not
// exported, we copy them below (and need to be consistent with the runtime).
func FindFuncWithName(name string) (uintptr, error) {
	for _, moduleData := range activeModules() {
		for _, ftab := range moduleData.ftab {
			f := (*runtime.Func)(unsafe.Pointer(&moduleData.pclntable[ftab.funcoff]))
			if getName(f) == name {
				return f.Entry(), nil
			}
		}
	}
	return 0, fmt.Errorf("invalid function name: %s", name)
}

func getName(f *runtime.Func) string {
	defer func() {
		recover()
	}()
	return f.Name()
}

//go:linkname activeModules runtime.activeModules
func activeModules() []*moduledata

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
	// Make sure moduledata is consistent with runtime.moduledata.
	pclntableField, _ := reflect.TypeOf(moduledata{}).FieldByName("pclntable")
	ftabField, _ := reflect.TypeOf(moduledata{}).FieldByName("ftab")
	rtModuledataTyp := reflect.TypeOf(activeModules()[0]).Elem()
	rt_pclntableField, ok1 := rtModuledataTyp.FieldByName("pclntable")
	rt_ftabField, ok2 := rtModuledataTyp.FieldByName("ftab")
	if !ok1 || !ok2 ||
		pclntableField.Offset != rt_pclntableField.Offset ||
		ftabField.Offset != rt_ftabField.Offset {
		panic("forceexport: moduledata structure not match")
	}
}
