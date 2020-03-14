package easy

import (
	"fmt"
	"reflect"
	"sync"
)

type SafeMap struct {
	sync.RWMutex
	Map map[interface{}]interface{}
}

func NewSafeMap() *SafeMap {
	return &SafeMap{Map: make(map[interface{}]interface{})}
}

type SafeInt64Map struct {
	sync.RWMutex
	Map map[int64]interface{}
}

func NewSafeInt64sMap() *SafeInt64Map {
	return &SafeInt64Map{Map: make(map[int64]interface{})}
}

type SafeStringMap struct {
	sync.RWMutex
	Map map[string]interface{}
}

func NewSafeStringMap() *SafeStringMap {
	return &SafeStringMap{Map: make(map[string]interface{})}
}

func MapKeys(m interface{}) (keys interface{}) {
	mTyp := reflect.TypeOf(m)
	if mTyp.Kind() != reflect.Map {
		panic(fmt.Sprintf("MapKeys: invalid type %T", m))
	}

	//return _iterMapKeys_reflect(m)
	return _iterMapKeys_unsafe(m)
}

func _iterMapKeys_reflect(m interface{}) interface{} {
	mTyp := reflect.TypeOf(m)
	mVal := reflect.ValueOf(m)
	keysVal := reflect.MakeSlice(reflect.SliceOf(mTyp.Key()), 0, mVal.Len())
	keysVal = reflect.Append(keysVal, mVal.MapKeys()...)
	return keysVal.Interface()
}

func MapValues(m interface{}) (values interface{}) {
	mTyp := reflect.TypeOf(m)
	if mTyp.Kind() != reflect.Map {
		panic(fmt.Sprintf("MapValues: invalid type %T", m))
	}

	//return _iterMapValues_reflect(m)
	return _iterMapValues_unsafe(m)
}

func _iterMapValues_reflect(m interface{}) interface{} {
	mTyp := reflect.TypeOf(m)
	mVal := reflect.ValueOf(m)
	valuesVal := reflect.MakeSlice(reflect.SliceOf(mTyp.Elem()), 0, mVal.Len())
	for iter := mVal.MapRange(); iter.Next(); {
		valuesVal = reflect.Append(valuesVal, iter.Value())
	}
	return valuesVal.Interface()
}

func IntKeys(m interface{}) (keys Int64s) {
	mTyp := reflect.TypeOf(m)
	if mTyp.Kind() != reflect.Map || !isIntType(mTyp.Key().Kind()) {
		panic(fmt.Sprintf("IntKeys: invalid type %T", m))
	}

	return _iterIntKeys(mTyp.Key().Kind(), m)
}

func IntValues(m interface{}) (values Int64s) {
	mTyp := reflect.TypeOf(m)
	if mTyp.Kind() != reflect.Map || !isIntType(mTyp.Elem().Kind()) {
		panic(fmt.Sprintf("IntValues: invalid type %T", m))
	}

	return _iterIntValues(mTyp.Elem().Kind(), m)
}

func StringKeys(m interface{}) (keys Strings) {
	mTyp := reflect.TypeOf(m)
	if mTyp.Kind() != reflect.Map || mTyp.Key().Kind() != reflect.String {
		panic(fmt.Sprintf("StringKeys: invalid type %T", m))
	}

	return _iterStringKeys(m)
}

func StringValues(m interface{}) (values Strings) {
	mTyp := reflect.TypeOf(m)
	if mTyp.Kind() != reflect.Map || mTyp.Elem().Kind() != reflect.String {
		panic(fmt.Sprintf("StringValues: invalid type %T", m))
	}

	return _iterStringValues(m)
}
