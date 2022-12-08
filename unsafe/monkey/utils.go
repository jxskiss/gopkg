package monkey

import (
	"fmt"
	"reflect"
	"syscall"
	"unsafe"
)

var _PAGE_SIZE = uintptr(syscall.Getpagesize())

func pageStart(ptr uintptr) uintptr {
	return ptr & ^(_PAGE_SIZE - 1)
}

// type value is a copy of reflect.Value.
type value struct {
	typ  unsafe.Pointer
	ptr  unsafe.Pointer
	flag uintptr
}

func getPtr(v reflect.Value) unsafe.Pointer {
	return (*value)(unsafe.Pointer(&v)).ptr
}

func getCode(target uintptr, length int) []byte {
	return *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: target,
		Len:  length,
		Cap:  length,
	}))
}

func slicePtr(b []byte) uintptr {
	return (*reflect.SliceHeader)(unsafe.Pointer(&b)).Data
}

func assertFunc(target interface{}, argName string) {
	if reflect.TypeOf(target).Kind() != reflect.Func {
		panic(fmt.Sprintf("monkey: want a function for %s but got %T", argName, target))
	}
}

func assertSameFuncType(target, repl interface{}) {
	targetTyp := reflect.TypeOf(target)
	replTyp := reflect.TypeOf(repl)
	if targetTyp.Kind() != reflect.Func {
		panic(fmt.Sprintf("monkey: target %v is not a function", target))
	}
	if replTyp.Kind() != reflect.Func {
		panic(fmt.Sprintf("monkey: replacement %v is not a function", repl))
	}
	if targetTyp != replTyp {
		panic(fmt.Sprintf("monkey: target and replacement have different type"))
	}
}

func assertReturnTypes(target reflect.Value, rets []interface{}) {
	if !target.IsValid() {
		panic("monkey: need a valid target to mock")
	}
	targetTyp := target.Type()
	if targetTyp.Kind() != reflect.Func {
		panic("monkey: target is not a function")
	}
	if targetTyp.NumOut() != len(rets) {
		panic(fmt.Sprintf("monkey: return values length not match, %d != %d",
			targetTyp.NumOut(), len(rets)))
	}
	for i := 0; i < targetTyp.NumOut(); i++ {
		if rets[i] == nil {
			continue
		}
		retTyp := reflect.TypeOf(rets[i])
		outTyp := targetTyp.Out(i)
		if !retTyp.ConvertibleTo(outTyp) {
			panic(fmt.Sprintf("monkey: return value type not match, %v != %v", retTyp, outTyp))
		}
	}
}

func assertVarPtr(targetAddr interface{}) {
	if reflect.TypeOf(targetAddr).Kind() != reflect.Ptr {
		panic("monkey: targetAddr is not a pointer to a variable")
	}
}

func assertVarReplacement(targetAddr, repl reflect.Value) {
	if targetAddr.Type().Kind() != reflect.Ptr {
		panic("monkey: targetAddr is not a pointer to a variable")
	}
	targetTyp := targetAddr.Type().Elem()
	replTyp := repl.Type()
	if !replTyp.ConvertibleTo(targetTyp) {
		panic(fmt.Sprintf("monkey: replacement %v can not be set to target %v", replTyp, targetTyp))
	}
}
