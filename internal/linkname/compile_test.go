package linkname

import (
	"reflect"
	"testing"
)

func TestCompile(t *testing.T) {
	compileReflectFunctions()
	compileRuntimeFunctions()
}

// call helps to ensure the linked functions can build.
func call(f any) {
	defer func() {
		recover()
	}()
	reflect.ValueOf(f).Call([]reflect.Value{})
}
