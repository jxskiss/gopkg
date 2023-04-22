package monkey

import (
	"fmt"
	"reflect"

	"github.com/jxskiss/gopkg/v2/unsafe/forceexport"
)

// Mock returns a Mocker object which helps to do mocking.
func Mock(target ...any) *Mocker {
	m := &Mocker{}
	if len(target) > 0 {
		m.Target(target[0])
	}
	return m
}

type Mocker struct {
	target reflect.Value
	repl   reflect.Value
	byName string
}

// Target sets the target to mock.
func (m *Mocker) Target(target any) *Mocker {
	assertFunc(target, "target")
	m.target = reflect.ValueOf(target)
	return m
}

// Method sets a method of a type as the mocking target.
func (m *Mocker) Method(target any, method string) *Mocker {
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
func (m *Mocker) ByName(name string, signature any) *Mocker {
	m.byName = name
	targetPtr := forceexport.FindFuncWithName(name)
	targetTyp := reflect.TypeOf(signature)
	targetVal := reflect.New(targetTyp)
	forceexport.CreateFuncForCodePtr(targetVal.Interface(), targetPtr)
	m.target = targetVal.Elem()
	return m
}

// Return sets the patch to build a function as replacement which returns rets.
func (m *Mocker) Return(rets ...any) *Mocker {
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
func (m *Mocker) To(repl any) *Mocker {
	if !m.target.IsValid() {
		panic("monkey: need a valid target to mock")
	}
	assertSameFuncType(m.target.Interface(), repl)
	return m.setReplacement(reflect.ValueOf(repl))
}

func (m *Mocker) setReplacement(repl reflect.Value) *Mocker {
	m.repl = repl
	return m
}

// Build applies the patch, it returns the final Patch object.
func (m *Mocker) Build() *Patch {
	if !m.target.IsValid() {
		panic("monkey: need a valid target to mock")
	}
	if !m.repl.IsValid() {
		panic("monkey: need a valid replacement to mock")
	}
	patch := patchFunc(m.target, m.repl)
	return patch
}
