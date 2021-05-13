package linkname

import (
	"reflect"
	"testing"
)

func TestCompile(t *testing.T) {
	compileReflectFunctions()
	compileRuntimeFunctions()
	compileUtilFunctions()
}

// call helps to ensure the linked functions can build.
func call(f interface{}) {
	defer func() {
		recover()
	}()
	reflect.ValueOf(f).Call([]reflect.Value{})
}
