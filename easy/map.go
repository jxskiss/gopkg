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

	mVal := reflect.ValueOf(m)
	keysVal := reflect.MakeSlice(reflect.SliceOf(mTyp.Key()), 0, mVal.Len())
	for _, k := range mVal.MapKeys() {
		keysVal = reflect.Append(keysVal, k)
	}
	return keysVal.Interface()
}

func MapValues(m interface{}) (values interface{}) {
	mTyp := reflect.TypeOf(m)
	if mTyp.Kind() != reflect.Map {
		panic(fmt.Sprintf("MapValues: invalid type %T", m))
	}

	mVal := reflect.ValueOf(m)
	valuesVal := reflect.MakeSlice(reflect.SliceOf(mTyp.Elem()), 0, mVal.Len())
	for _, k := range mVal.MapKeys() {
		valuesVal = reflect.Append(valuesVal, mVal.MapIndex(k))
	}
	return valuesVal.Interface()
}

func IntKeys(m interface{}) (keys Int64s) {
	mTyp := reflect.TypeOf(m)
	if mTyp.Kind() != reflect.Map || !isIntType(mTyp.Key()) {
		panic(fmt.Sprintf("IntKeys: invalid type %T", m))
	}

	mVal := reflect.ValueOf(m)
	keys = make([]int64, mVal.Len())
	for i, k := range mVal.MapKeys() {
		keys[i] = reflectInt(k)
	}
	return
}

func IntValues(m interface{}) (values Int64s) {
	mTyp := reflect.TypeOf(m)
	if mTyp.Kind() != reflect.Map || !isIntType(mTyp.Elem()) {
		panic(fmt.Sprintf("IntValues: invalid type %T", m))
	}

	mVal := reflect.ValueOf(m)
	values = make([]int64, mVal.Len())
	for i, k := range mVal.MapKeys() {
		values[i] = reflectInt(mVal.MapIndex(k))
	}
	return
}

func StringKeys(m interface{}) (keys []string) {
	mTyp := reflect.TypeOf(m)
	if mTyp.Kind() != reflect.Map || mTyp.Key().Kind() != reflect.String {
		panic(fmt.Sprintf("StringKeys: unsupporteeed type %T", m))
	}

	mVal := reflect.ValueOf(m)
	keys = make([]string, mVal.Len())
	for i, k := range mVal.MapKeys() {
		keys[i] = k.String()
	}
	return
}

func StringValues(m interface{}) (values []string) {
	mTyp := reflect.TypeOf(m)
	if mTyp.Kind() != reflect.Map || mTyp.Elem().Kind() != reflect.String {
		panic(fmt.Sprintf("StringValues: invalid type %T", m))
	}

	mVal := reflect.ValueOf(m)
	values = make([]string, mVal.Len())
	for i, k := range mVal.MapKeys() {
		values[i] = mVal.MapIndex(k).String()
	}
	return
}
