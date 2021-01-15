package easy

import (
	"fmt"
	"github.com/jxskiss/gopkg/reflectx"
	"reflect"
	"sync"
	"time"
	"unsafe"
)

func MapKeys(m interface{}) (keys interface{}) {
	mTyp := reflect.TypeOf(m)
	if mTyp.Kind() != reflect.Map {
		panic(invalidType("MapKeys", "map", m))
	}

	length := reflectx.MapLen(m)
	keyTyp := mTyp.Key()
	keySize := keyTyp.Size()
	slice, header, keyRType := reflectx.MakeSlice(keyTyp, length, length)
	array := header.Data
	i := 0
	reflectx.MapIter(m, func(k, _ unsafe.Pointer) {
		dst := reflectx.ArrayAt(array, i, keySize)
		reflectx.TypedMemMove(keyRType, dst, k)
		i++
	})
	return slice
}

func MapValues(m interface{}) (values interface{}) {
	mTyp := reflect.TypeOf(m)
	if mTyp.Kind() != reflect.Map {
		panic(invalidType("MapValues", "map", m))
	}

	length := reflectx.MapLen(m)
	elemTyp := mTyp.Elem()
	elemSize := elemTyp.Size()
	slice, header, elemRType := reflectx.MakeSlice(elemTyp, length, length)
	array := header.Data
	i := 0
	reflectx.MapIter(m, func(_, v unsafe.Pointer) {
		dst := reflectx.ArrayAt(array, i, elemSize)
		reflectx.TypedMemMove(elemRType, dst, v)
		i++
	})
	return slice
}

func IntKeys(m interface{}) (keys Int64s) {
	mTyp := reflect.TypeOf(m)
	if mTyp.Kind() != reflect.Map ||
		!reflectx.IsIntType(mTyp.Key().Kind()) {
		panic(invalidType("IntKeys", "map with integer keys", m))
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
		panic(invalidType("IntValues", "map with integer values", m))
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
		panic(invalidType("StringKeys", "map with string keys", m))
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
		panic(invalidType("StringValues", "map with string values", m))
	}

	out := make([]string, 0, reflectx.MapLen(m))
	reflectx.MapIter(m, func(_, v unsafe.Pointer) {
		x := *(*string)(v)
		out = append(out, x)
	})
	return out
}

func MergeMap(map1, map2 interface{}) interface{} {
	if map1 == nil || map2 == nil {
		panicNilParams("MergeMap", "map1", map1, "map2", map2)
	}
	m1Val := reflect.ValueOf(map1)
	m2Val := reflect.ValueOf(map2)
	if m1Val.Type() != m2Val.Type() || m1Val.Kind() != reflect.Map {
		panic("MergeMap: " + errNotSameTypeOrNotMap)
	}

	for iter := m2Val.MapRange(); iter.Next(); {
		m1Val.SetMapIndex(iter.Key(), iter.Value())
	}
	return m1Val.Interface()
}

type SafeMap struct {
	sync.RWMutex
	Map Map
}

func NewSafeMap() *SafeMap {
	return &SafeMap{Map: make(Map)}
}

type SafeInt64Map struct {
	sync.RWMutex
	Map Int64Map
}

func NewSafeInt64Map() *SafeInt64Map {
	return &SafeInt64Map{Map: make(Int64Map)}
}

type Int64Map map[int64]interface{}

type Map map[string]interface{}

// Set is used to store a new key/value pair exclusively in the map.
// It also lazy initializes the map if it was not used previously.
func (p *Map) Set(key string, value interface{}) {
	if *p == nil {
		*p = make(Map)
	}
	(*p)[key] = value
}

// Get returns the value for the given key, ie: (value, true).
// If the value does not exists it returns (nil, false)
func (p Map) Get(key string) (value interface{}, exists bool) {
	value, exists = p[key]
	return
}

// MustGet returns the value for the given key if it exists, otherwise it panics.
func (p Map) MustGet(key string) interface{} {
	if val, ok := p[key]; ok {
		return val
	}
	panic(fmt.Sprintf("key %q not exists", key))
}

// GetString returns the value associated with the key as a string.
func (p Map) GetString(key string) string {
	if val, ok := p[key].(string); ok {
		return val
	}
	return ""
}

// GetBool returns the value associated with the key as a boolean.
func (p Map) GetBool(key string) bool {
	if val, ok := p[key].(bool); ok {
		return val
	}
	return false
}

// GetInt returns the value associated with the key as an integer.
func (p Map) GetInt(key string) int {
	if val, ok := p[key].(int); ok {
		return val
	}
	return 0
}

// GetInt64 returns the value associated with the key as an int64.
func (p Map) GetInt64(key string) int64 {
	if val, ok := p[key].(int64); ok {
		return val
	}
	return 0
}

// GetInt32 returns the value associated with the key as an int32.
func (p Map) GetInt32(key string) int32 {
	if val, ok := p[key].(int32); ok {
		return val
	}
	return 0
}

// GetFloat64 returns the value associated with the key as a float64.
func (p Map) GetFloat64(key string) float64 {
	if val, ok := p[key].(float64); ok {
		return val
	}
	return 0
}

// GetTime returns the value associated with the key as time.
func (p Map) GetTime(key string) time.Time {
	if val, ok := p[key].(time.Time); ok {
		return val
	}
	return time.Time{}
}

// GetDuration returns the value associated with the key as a duration.
func (p Map) GetDuration(key string) time.Duration {
	if val, ok := p[key].(time.Duration); ok {
		return val
	}
	return 0
}

// GetInt64s returns the value associated with the key as a slice of int64.
func (p Map) GetInt64s(key string) Int64s {
	val, ok := p[key]
	if ok {
		switch val := val.(type) {
		case Int64s:
			return val
		case []int64:
			return val
		}
	}
	return nil
}

// GetInt32s returns the value associated with the key as a slice of int32.
func (p Map) GetInt32s(key string) Int32s {
	val, ok := p[key]
	if ok {
		switch val := val.(type) {
		case Int32s:
			return val
		case []int32:
			return val
		}
	}
	return nil
}

// GetStrings returns the value associated with the key as a slice of strings.
func (p Map) GetStrings(key string) Strings {
	val, ok := p[key]
	if ok {
		switch val := val.(type) {
		case Strings:
			return val
		case []string:
			return val
		}
	}
	return nil
}

// GetMap returns the value associated with the key as a Map (map[string]interface{}).
func (p Map) GetMap(key string) Map {
	val, ok := p[key]
	if ok {
		switch val := val.(type) {
		case Map:
			return val
		case map[string]interface{}:
			return val
		}
	}
	return nil
}

// GetInt64Map returns the value associated with the key as an Int64Map (map[int64]interface{}).
func (p Map) GetInt64Map(key string) Int64Map {
	val, ok := p[key]
	if ok {
		switch val := val.(type) {
		case Int64Map:
			return val
		case map[int64]interface{}:
			return val
		}
	}
	return nil
}

// GetStringMap returns the value associated with the key as a map of strings (map[string]string).
func (p Map) GetStringMap(key string) map[string]string {
	if val, ok := p[key].(map[string]string); ok {
		return val
	}
	return nil
}

// Iterate iterates the map in unspecified order, the given function fn
// will be called for each key value pair.
func (p Map) Iterate(fn func(k string, v interface{})) {
	for k, v := range p {
		fn(k, v)
	}
}

// Set is used to store a new key/value pair exclusively in the map.
// It also lazy initializes the map if it was not used previously.
func (p *Int64Map) Set(key int64, value interface{}) {
	if *p == nil {
		*p = make(Int64Map)
	}
	(*p)[key] = value
}

// Get returns the value for the given key, ie: (value, true).
// If the value does not exists it returns (nil, false)
func (p Int64Map) Get(key int64) (value interface{}, exists bool) {
	value, exists = p[key]
	return
}

// MustGet returns the value for the given key if it exists, otherwise it panics.
func (p Int64Map) MustGet(key int64) interface{} {
	if val, ok := p[key]; ok {
		return val
	}
	panic(fmt.Sprintf("key %q not exists", key))
}

// GetString returns the value associated with the key as a string.
func (p Int64Map) GetString(key int64) string {
	if val, ok := p[key].(string); ok {
		return val
	}
	return ""
}

// GetBool returns the value associated with the key as a boolean.
func (p Int64Map) GetBool(key int64) bool {
	if val, ok := p[key].(bool); ok {
		return val
	}
	return false
}

// GetInt returns the value associated with the key as an integer.
func (p Int64Map) GetInt(key int64) int {
	if val, ok := p[key].(int); ok {
		return val
	}
	return 0
}

// GetInt64 returns the value associated with the key as an int64.
func (p Int64Map) GetInt64(key int64) int64 {
	if val, ok := p[key].(int64); ok {
		return val
	}
	return 0
}

// GetInt32 returns the value associated with the key as an int32.
func (p Int64Map) GetInt32(key int64) int32 {
	if val, ok := p[key].(int32); ok {
		return val
	}
	return 0
}

// GetFloat64 returns the value associated with the key as a float64.
func (p Int64Map) GetFloat64(key int64) float64 {
	if val, ok := p[key].(float64); ok {
		return val
	}
	return 0
}

// GetTime returns the value associated with the key as time.
func (p Int64Map) GetTime(key int64) time.Time {
	if val, ok := p[key].(time.Time); ok {
		return val
	}
	return time.Time{}
}

// GetDuration returns the value associated with the key as a duration.
func (p Int64Map) GetDuration(key int64) time.Duration {
	if val, ok := p[key].(time.Duration); ok {
		return val
	}
	return 0
}

// GetInt64s returns the value associated with the key as a slice of int64.
func (p Int64Map) GetInt64s(key int64) Int64s {
	val, ok := p[key]
	if ok {
		switch val := val.(type) {
		case Int64s:
			return val
		case []int64:
			return val
		}
	}
	return nil
}

// GetInt32s returns the value associated with the key as a slice of int32.
func (p Int64Map) GetInt32s(key int64) Int32s {
	val, ok := p[key]
	if ok {
		switch val := val.(type) {
		case Int32s:
			return val
		case []int32:
			return val
		}
	}
	return nil
}

// GetStrings returns the value associated with the key as a slice of strings.
func (p Int64Map) GetStrings(key int64) Strings {
	val, ok := p[key]
	if ok {
		switch val := val.(type) {
		case Strings:
			return val
		case []string:
			return val
		}
	}
	return nil
}

// GetMap returns the value associated with the key as a Map (map[string]interface{}).
func (p Int64Map) GetMap(key int64) Map {
	val, ok := p[key]
	if ok {
		switch val := val.(type) {
		case Map:
			return val
		case map[string]interface{}:
			return val
		}
	}
	return nil
}

// GetInt64Map returns the value associated with the key as an Int64Map (map[int64]interface{}).
func (p Int64Map) GetInt64Map(key int64) Int64Map {
	val, ok := p[key]
	if ok {
		switch val := val.(type) {
		case Int64Map:
			return val
		case map[int64]interface{}:
			return val
		}
	}
	return nil
}

// GetStringMap returns the value associated with the key as a map of strings (map[string]string).
func (p Int64Map) GetStringMap(key int64) map[string]string {
	if val, ok := p[key].(map[string]string); ok {
		return val
	}
	return nil
}

// Iterate iterates the map in unspecified order, the given function fn
// will be called for each key value pair.
func (p Int64Map) Iterate(fn func(k int64, v interface{})) {
	for k, v := range p {
		fn(k, v)
	}
}
