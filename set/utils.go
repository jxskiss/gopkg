package set

import (
	"fmt"
	"reflect"
)

// IntKeys returns int key slice of a map, the given map's key type
// must be int, or it will panic.
//
// For many frequently used types, type assertion is used to get best perf,
// else reflect is used to support any map type with int keys.
func IntKeys(m interface{}) (keys []int) {
	switch v := m.(type) {
	case map[int]string:
		keys = make([]int, len(v))
		i := 0
		for k := range v {
			keys[i] = k
			i++
		}
	case map[int][]byte:
		keys = make([]int, len(v))
		i := 0
		for k := range v {
			keys[i] = k
			i++
		}
	case map[int]int:
		keys = make([]int, len(v))
		i := 0
		for k := range v {
			keys[i] = k
			i++
		}
	case map[int]int64:
		keys = make([]int, len(v))
		i := 0
		for k := range v {
			keys[i] = k
			i++
		}
	case map[int]uint64:
		keys = make([]int, len(v))
		i := 0
		for k := range v {
			keys[i] = k
			i++
		}
	case map[int]bool:
		keys = make([]int, len(v))
		i := 0
		for k := range v {
			keys[i] = k
			i++
		}
	case map[int]struct{}:
		keys = make([]int, len(v))
		i := 0
		for k := range v {
			keys[i] = k
			i++
		}
	case map[int]interface{}:
		keys = make([]int, len(v))
		i := 0
		for k := range v {
			keys[i] = k
			i++
		}
	default:
		mTyp := reflect.TypeOf(m)
		if mTyp.Kind() != reflect.Map || !isIntType(mTyp.Key()) {
			panic(fmt.Errorf("unsupported type or IntKeys: %T", v))
		}

		mVal := reflect.ValueOf(m)
		keys = make([]int, mVal.Len())
		for i, k := range mVal.MapKeys() {
			keys[i] = int(reflectInt(k))
		}
	}
	return keys
}

// Int64Keys returns int64 key slice of a map, the given map's key type
// must be int64, or it will panic.
//
// For many frequently used types, type assertion is used to get best perf,
// else reflect is used to support any map type with int64 keys.
func Int64Keys(m interface{}) (keys []int64) {
	switch v := m.(type) {
	case map[int64]string:
		keys = make([]int64, len(v))
		i := 0
		for k := range v {
			keys[i] = k
			i++
		}
	case map[int64][]byte:
		keys = make([]int64, len(v))
		i := 0
		for k := range v {
			keys[i] = k
			i++
		}
	case map[int64]int:
		keys = make([]int64, len(v))
		i := 0
		for k := range v {
			keys[i] = k
			i++
		}
	case map[int64]int64:
		keys = make([]int64, len(v))
		i := 0
		for k := range v {
			keys[i] = k
			i++
		}
	case map[int64]uint64:
		keys = make([]int64, len(v))
		i := 0
		for k := range v {
			keys[i] = k
			i++
		}
	case map[int64]bool:
		keys = make([]int64, len(v))
		i := 0
		for k := range v {
			keys[i] = k
			i++
		}
	case map[int64]struct{}:
		keys = make([]int64, len(v))
		i := 0
		for k := range v {
			keys[i] = k
			i++
		}
	case map[int64]interface{}:
		keys = make([]int64, len(v))
		i := 0
		for k := range v {
			keys[i] = k
			i++
		}
	default:
		mTyp := reflect.TypeOf(m)
		if mTyp.Kind() != reflect.Map || !isIntType(mTyp.Key()) {
			panic(fmt.Errorf("unsupported type for Int64Keys: %T", v))
		}

		mVal := reflect.ValueOf(m)
		keys = make([]int64, mVal.Len())
		for i, k := range mVal.MapKeys() {
			keys[i] = reflectInt(k)
		}
	}
	return keys
}

func Uint64Keys(m interface{}) (keys []uint64) {
	switch v := m.(type) {
	case map[uint64]string:
		keys := make([]uint64, len(v))
		i := 0
		for k := range v {
			keys[i] = k
			i++
		}
	case map[uint64][]byte:
		keys := make([]uint64, len(v))
		i := 0
		for k := range v {
			keys[i] = k
			i++
		}
	case map[uint64]int:
		keys := make([]uint64, len(v))
		i := 0
		for k := range v {
			keys[i] = k
			i++
		}
	case map[uint64]int64:
		keys := make([]uint64, len(v))
		i := 0
		for k := range v {
			keys[i] = k
			i++
		}
	case map[uint64]uint64:
		keys := make([]uint64, len(v))
		i := 0
		for k := range v {
			keys[i] = k
			i++
		}
	case map[uint64]bool:
		keys := make([]uint64, len(v))
		i := 0
		for k := range v {
			keys[i] = k
			i++
		}
	case map[uint64]struct{}:
		keys := make([]uint64, len(v))
		i := 0
		for k := range v {
			keys[i] = k
			i++
		}
	case map[uint64]interface{}:
		keys := make([]uint64, len(v))
		i := 0
		for k := range v {
			keys[i] = k
			i++
		}
	default:
		mTyp := reflect.TypeOf(m)
		if mTyp.Kind() != reflect.Map || !isIntType(mTyp.Key()) {
			panic(fmt.Errorf("unsupported type for Int64Keys: %T", v))
		}

		mVal := reflect.ValueOf(m)
		keys := make([]uint64, mVal.Len())
		for i, k := range mVal.MapKeys() {
			keys[i] = uint64(reflectInt(k))
		}
	}
	return keys
}

func isIntType(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	default:
		return false
	}
}

func reflectInt(v reflect.Value) int64 {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return int64(v.Uint())
	}

	// shall not happen, type should be pre-checked
	panic(fmt.Errorf("reflectInt: not int type: %s", v.Kind().String()))
}

// StringKeys returns string key slice of a map, the given map's key type
// must be string, or it will panic.
//
// For many frequently used types, type assertion is used to get best perf,
// else reflect is used to support any map type with string keys.
func StringKeys(m interface{}) (keys []string) {
	switch v := m.(type) {
	case map[string]string:
		keys = make([]string, len(v))
		i := 0
		for k := range v {
			keys[i] = k
			i++
		}
	case map[string][]byte:
		keys = make([]string, len(v))
		i := 0
		for k := range v {
			keys[i] = k
			i++
		}
	case map[string]int:
		keys = make([]string, len(v))
		i := 0
		for k := range v {
			keys[i] = k
			i++
		}
	case map[string]int64:
		keys = make([]string, len(v))
		i := 0
		for k := range v {
			keys[i] = k
			i++
		}
	case map[string]uint64:
		keys = make([]string, len(v))
		i := 0
		for k := range v {
			keys[i] = k
			i++
		}
	case map[string]bool:
		keys = make([]string, len(v))
		i := 0
		for k := range v {
			keys[i] = k
			i++
		}
	case map[string]struct{}:
		keys = make([]string, len(v))
		i := 0
		for k := range v {
			keys[i] = k
			i++
		}
	case map[string]interface{}:
		keys = make([]string, len(v))
		i := 0
		for k := range v {
			keys[i] = k
			i++
		}
	default:
		mTyp := reflect.TypeOf(m)
		if mTyp.Kind() != reflect.Map || mTyp.Key().Kind() != reflect.String {
			panic(fmt.Errorf("unsupported type for StringKeys: %T", v))
		}

		mVal := reflect.ValueOf(m)
		keys = make([]string, 0, mVal.Len())
		for _, k := range mVal.MapKeys() {
			keys = append(keys, k.String())
		}
	}
	return keys
}
