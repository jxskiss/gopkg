package gemap

import (
	"fmt"
	"github.com/jxskiss/gopkg/v2/internal"
	"github.com/jxskiss/gopkg/v2/internal/linkname"
	"reflect"
)

// Keys is reserved for a generic implementation.

// Values is reserved for a generic implementation.

// MapKeys returns a slice containing all the keys present in the map,
// in unspecified order.
// It panics if m's kind is not reflect.Map.
// It returns an emtpy slice if m is a nil map.
func MapKeys(m interface{}) (keys interface{}) {
	mTyp := reflect.TypeOf(m)
	if mTyp.Kind() != reflect.Map {
		panic(invalidType("MapKeys", "map", m))
	}

	mVal := reflect.ValueOf(m)
	keyTyp := mTyp.Key()
	keySliceTyp := reflect.SliceOf(keyTyp)
	length := mVal.Len()
	keysVal := reflect.MakeSlice(keySliceTyp, 0, length)
	for _, kVal := range mVal.MapKeys() {
		keysVal = reflect.Append(keysVal, kVal)
	}
	return keysVal.Interface()
}

// MapValues returns a slice containing all the values present in the map,
// in unspecified order.
// It panics if m's kind is not reflect.Map.
// It returns an empty slice if m is a nil map.
func MapValues(m interface{}) (values interface{}) {
	mTyp := reflect.TypeOf(m)
	if mTyp.Kind() != reflect.Map {
		panic(invalidType("MapValues", "map", m))
	}

	mVal := reflect.ValueOf(m)
	elemTyp := mTyp.Elem()
	elemSliceTyp := reflect.SliceOf(elemTyp)
	length := mVal.Len()
	valuesVal := reflect.MakeSlice(elemSliceTyp, 0, length)
	for iter := mVal.MapRange(); iter.Next(); {
		valuesVal = reflect.Append(valuesVal, iter.Value())
	}
	return valuesVal.Interface()
}

// IntKeys returns a int64 slice containing all the keys present in the map,
// in unspecified order.
// It panics if m's kind is not reflect.Map or the key's type is not integer.
// It returns an empty slice if m is a nil map.
func IntKeys(m interface{}) (keys []int64) {
	mTyp := reflect.TypeOf(m)
	if mTyp.Kind() != reflect.Map || !isIntType(mTyp.Key().Kind()) {
		panic(invalidType("IntKeys", "map with integer keys", m))
	}

	eface := internal.EFaceOf(&m)
	length := linkname.Reflect_maplen(eface.Word)
	keyKind := mTyp.Key().Kind()
	iter := linkname.Reflect_mapiterinit(eface.RType, eface.Word)
	keys = make([]int64, 0, length)
	for i := 0; i < length; i++ {
		keyptr := linkname.Reflect_mapiterkey(iter)
		if keyptr == nil {
			break
		}
		keys = append(keys, internal.CastIntPointer(keyKind, keyptr))
		linkname.Reflect_mapiternext(iter)
	}
	return keys
}

// IntValues returns a int64 slice containing all the values present in the map,
// in unspecified order.
// It panics if m's kind is not reflect.Map or the value's type is not integer.
// It returns an empty slice if m is a nil map.
func IntValues(m interface{}) (values []int64) {
	mTyp := reflect.TypeOf(m)
	if mTyp.Kind() != reflect.Map || !isIntType(mTyp.Elem().Kind()) {
		panic(invalidType("IntValues", "map with integer values", m))
	}

	eface := internal.EFaceOf(&m)
	length := linkname.Reflect_maplen(eface.Word)
	elemKind := mTyp.Elem().Kind()
	iter := linkname.Reflect_mapiterinit(eface.RType, eface.Word)
	values = make([]int64, 0, length)
	for i := 0; i < length; i++ {
		keyptr := linkname.Reflect_mapiterkey(iter)
		if keyptr == nil {
			break
		}
		elemptr := linkname.Reflect_mapiterelem(iter)
		values = append(values, internal.CastIntPointer(elemKind, elemptr))
		linkname.Reflect_mapiternext(iter)
	}
	return values
}

// StringKeys returns a string slice containing all the keys present
// in the map, in unspecified order.
// It panics if m's kind is not reflect.Map or the key's kind is not string.
// It returns an empty slice if m is a nil map.
func StringKeys(m interface{}) (keys []string) {
	mTyp := reflect.TypeOf(m)
	if mTyp.Kind() != reflect.Map || mTyp.Key().Kind() != reflect.String {
		panic(invalidType("StringKeys", "map with string keys", m))
	}

	eface := internal.EFaceOf(&m)
	length := linkname.Reflect_maplen(eface.Word)
	iter := linkname.Reflect_mapiterinit(eface.RType, eface.Word)
	keys = make([]string, 0, length)
	for i := 0; i < length; i++ {
		keyptr := linkname.Reflect_mapiterkey(iter)
		if keyptr == nil {
			break
		}
		keys = append(keys, *(*string)(keyptr))
		linkname.Reflect_mapiternext(iter)
	}
	return keys
}

// StringValues returns a string slice containing all the values present
// in the map, in unspecified order.
// It panics if m's kind is not reflect.Map or the value's kind is not string.
// It returns an empty slice if m is a nil map.
func StringValues(m interface{}) (values []string) {
	mTyp := reflect.TypeOf(m)
	if mTyp.Kind() != reflect.Map || mTyp.Elem().Kind() != reflect.String {
		panic(invalidType("StringValues", "map with string values", m))
	}

	eface := internal.EFaceOf(&m)
	length := linkname.Reflect_maplen(eface.Word)
	iter := linkname.Reflect_mapiterinit(eface.RType, eface.Word)
	values = make([]string, 0, length)
	for i := 0; i < length; i++ {
		keyptr := linkname.Reflect_mapiterkey(iter)
		if keyptr == nil {
			break
		}
		elemptr := linkname.Reflect_mapiterelem(iter)
		values = append(values, *(*string)(elemptr))
		linkname.Reflect_mapiternext(iter)
	}
	return values
}

// Merge and is reserved for a generic implementation.

// MergeTo is reserved for a generic implementation.

// MergeMaps returns a new map containing all key values present in maps.
// It panics if given zero param.
// It panics if given param which is not a map, or different map types.
// It returns an empty map if all params are nil or empty.
func MergeMaps(maps ...interface{}) interface{} {
	if len(maps) == 0 {
		panic(invalidParam("MergeMaps", "maps"))
	}
	var dstTyp reflect.Type
	var length int
	for _, m := range maps {
		if m == nil {
			continue
		}
		mTyp := reflect.TypeOf(m)
		if mTyp.Kind() != reflect.Map {
			panic(invalidType("MergeMaps", "map", m))
		}
		if dstTyp == nil {
			dstTyp = mTyp
			continue
		}
		if mTyp != dstTyp {
			panic(invalidType("MergeMaps", dstTyp.String(), m))
		}
		eface := internal.EFaceOf(&m)
		length += linkname.Reflect_maplen(eface.Word)
	}
	dstMap := reflect.MakeMapWithSize(dstTyp, length).Interface()
	return mergeMapsTo(dstMap, maps...)
}

// MergeMapsTo adds key values present in others to the dst map.
// It panics if dst is nil or not a map, or any param in others is not a map,
// or they are different map types.
// If dst is a nil map, it creates a new map and returns it.
func MergeMapsTo(dst interface{}, others ...interface{}) interface{} {
	dstTyp := reflect.TypeOf(dst)
	if dstTyp.Kind() != reflect.Map {
		panic(invalidType("MergeMapsTo", "map", dst))
	}
	for _, m := range others {
		if m == nil {
			continue
		}
		if reflect.TypeOf(m) != dstTyp {
			panic(invalidType("MergeMapsTo", dstTyp.String(), m))
		}
	}
	return mergeMapsTo(dst, others...)
}

func mergeMapsTo(dst interface{}, others ...interface{}) interface{} {
	dstVal := reflect.ValueOf(dst)
	if dstVal.IsNil() {
		dstTyp := reflect.TypeOf(dst)
		dstVal = reflect.MakeMap(dstTyp)
	}
	for _, m := range others {
		if m == nil {
			continue
		}
		mVal := reflect.ValueOf(m)
		for iter := mVal.MapRange(); iter.Next(); {
			dstVal.SetMapIndex(iter.Key(), iter.Value())
		}
	}
	return dstVal.Interface()
}

func invalidType(where string, want string, got interface{}) string {
	const invalidType = "%s: invalid type, want %s, got %T"
	return fmt.Sprintf(invalidType, where, want, got)
}

func invalidParam(where string, name string) string {
	const invalidParam = "%s: invalid param %s"
	return fmt.Sprintf(invalidParam, where, name)
}

func isIntType(kind reflect.Kind) bool {
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return true
	}
	return false
}
