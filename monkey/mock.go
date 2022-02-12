package monkey

import (
	"fmt"
	"reflect"

	"github.com/jxskiss/gopkg/v2/forceexport"
)

// Mock returns a mock object which helps to do mocking.
func Mock() *mock {
	return &mock{}
}

type mock struct {
	target reflect.Value
	repl   reflect.Value
	byName string
}

// Target sets the target to mock.
func (m *mock) Target(target interface{}) *mock {
	assertFunc(target, "target")
	m.target = reflect.ValueOf(target)
	return m
}

// Method sets a method of a type as the mocking target.
func (m *mock) Method(target interface{}, method string) *mock {
	targetTyp := reflect.TypeOf(target)
	targetMethod, ok := targetTyp.MethodByName(method)
	if !ok {
		panic(fmt.Sprintf("monkey: unknown method %s.%s", targetTyp.Name(), method))
	}
	m.target = targetMethod.Func
	return m
}

// ByName sets the mocking target by name.
// Private method is supported by specifying the full name.
func (m *mock) ByName(name string, signature interface{}) *mock {
	m.byName = name
	targetPtr := forceexport.FindFuncWithName(name)
	targetTyp := reflect.TypeOf(signature)
	targetVal := reflect.New(targetTyp)
	forceexport.CreateFuncForCodePtr(targetVal.Interface(), targetPtr)
	m.target = targetVal.Elem()
	return m
}

// Return sets the patch to build a function as replacement which returns rets.
func (m *mock) Return(rets ...interface{}) *mock {
	if !m.target.IsValid() {
		panic("monkey: need a valid target to mock")
	}
	assertReturnTypes(m.target, rets)
	targetTyp := m.target.Type()
	repl := reflect.MakeFunc(targetTyp, func(args []reflect.Value) (results []reflect.Value) {
		for i, x := range rets {
			val := reflect.Zero(targetTyp.Out(i))
			if x != nil {
				val = reflect.ValueOf(x).Convert(targetTyp.Out(i))
			}
			results = append(results, val)
		}
		return
	})
	return m.setReplacement(repl)
}

// To sets the replacement to mock with.
func (m *mock) To(repl interface{}) *mock {
	if !m.target.IsValid() {
		panic("monkey: need a valid target to mock")
	}
	assertSameFuncType(m.target.Interface(), repl)
	return m.setReplacement(reflect.ValueOf(repl))
}

func (m *mock) setReplacement(repl reflect.Value) *mock {
	m.repl = repl
	return m
}

// Build applies the patch, it returns the final Patch object.
func (m *mock) Build() *Patch {
	if !m.target.IsValid() {
		panic("monkey: need a valid target to mock")
	}
	if !m.repl.IsValid() {
		panic("monkey: need a valid replacement to mock")
	}
	patch := patchFunc(m.target, m.repl)
	return patch
}
