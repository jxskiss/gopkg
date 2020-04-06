package easy

import (
	"fmt"
	"github.com/jxskiss/gopkg/reflectx"
	"reflect"
	"sync"
	"unsafe"
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

	length := reflectx.MapLen(m)
	keyTyp := mTyp.Key()
	keySize := keyTyp.Size()
	out, slice, keyRType := reflectx.MakeSlice(keyTyp, length, length)
	array := slice.Data
	i := 0
	reflectx.MapIter(m, func(k, _ unsafe.Pointer) {
		dst := reflectx.ArrayAt(array, i, keySize)
		reflectx.TypedMemMove(keyRType, dst, k)
		i++
	})
	return out
}

func MapValues(m interface{}) (values interface{}) {
	mTyp := reflect.TypeOf(m)
	if mTyp.Kind() != reflect.Map {
		panic(fmt.Sprintf("MapValues: invalid type %T", m))
	}

	length := reflectx.MapLen(m)
	elemTyp := mTyp.Elem()
	elemSize := elemTyp.Size()
	out, slice, elemRType := reflectx.MakeSlice(elemTyp, length, length)
	array := slice.Data
	i := 0
	reflectx.MapIter(m, func(_, v unsafe.Pointer) {
		dst := reflectx.ArrayAt(array, i, elemSize)
		reflectx.TypedMemMove(elemRType, dst, v)
		i++
	})
	return out
}

func IntKeys(m interface{}) (keys Int64s) {
	mTyp := reflect.TypeOf(m)
	if mTyp.Kind() != reflect.Map ||
		!reflectx.IsIntType(mTyp.Key().Kind()) {
		panic(fmt.Sprintf("IntKeys: invalid type %T", m))
	}

	out := make([]int64, 0, reflectx.MapLen(m))
	cast := reflectx.GetIntCaster(mTyp.Key().Kind()).Cast
	reflectx.MapIter(m, func(k, _ unsafe.Pointer) {
		out = append(out, cast(k))
	})
	return out
}

func IntValues(m interface{}) (values Int64s) {
	mTyp := reflect.TypeOf(m)
	if mTyp.Kind() != reflect.Map ||
		!reflectx.IsIntType(mTyp.Elem().Kind()) {
		panic(fmt.Sprintf("IntValues: invalid type %T", m))
	}

	out := make([]int64, 0, reflectx.MapLen(m))
	cast := reflectx.GetIntCaster(mTyp.Elem().Kind()).Cast
	reflectx.MapIter(m, func(_, v unsafe.Pointer) {
		out = append(out, cast(v))
	})
	return out
}

func StringKeys(m interface{}) (keys Strings) {
	mTyp := reflect.TypeOf(m)
	if mTyp.Kind() != reflect.Map || mTyp.Key().Kind() != reflect.String {
		panic(fmt.Sprintf("StringKeys: invalid type %T", m))
	}

	out := make([]string, 0, reflectx.MapLen(m))
	reflectx.MapIter(m, func(k, _ unsafe.Pointer) {
		x := *(*string)(k)
		out = append(out, x)
	})
	return out
}

func StringValues(m interface{}) (values Strings) {
	mTyp := reflect.TypeOf(m)
	if mTyp.Kind() != reflect.Map || mTyp.Elem().Kind() != reflect.String {
		panic(fmt.Sprintf("StringValues: invalid type %T", m))
	}

	out := make([]string, 0, reflectx.MapLen(m))
	reflectx.MapIter(m, func(_, v unsafe.Pointer) {
		x := *(*string)(v)
		out = append(out, x)
	})
	return out
}
