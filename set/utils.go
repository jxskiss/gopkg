package set

import (
	"fmt"
	"reflect"
)

func Int64Keys(m interface{}) (keys []int64) {
	switch v := m.(type) {
	case map[int64]string:
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
	case map[int64]int:
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
	case map[int64]map[int64]bool:
		keys = make([]int64, len(v))
		i := 0
		for k := range v {
			keys[i] = k
			i++
		}
	case map[int64]map[string]bool:
		keys = make([]int64, len(v))
		i := 0
		for k := range v {
			keys[i] = k
			i++
		}
	case map[int64][]int64:
		keys = make([]int64, len(v))
		i := 0
		for k := range v {
			keys[i] = k
			i++
		}
	case map[int64][]string:
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

func StringKeys(m interface{}) (keys []string) {
	switch v := m.(type) {
	case map[string]string:
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
	case map[string]int:
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
	case map[string]map[int64]bool:
		keys = make([]string, len(v))
		i := 0
		for k := range v {
			keys[i] = k
			i++
		}
	case map[string]map[string]bool:
		keys = make([]string, len(v))
		i := 0
		for k := range v {
			keys[i] = k
			i++
		}
	case map[string][]int64:
		keys = make([]string, len(v))
		i := 0
		for k := range v {
			keys[i] = k
			i++
		}
	case map[string][]string:
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
