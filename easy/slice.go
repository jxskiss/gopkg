package easy

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
)

var (
	ErrNotSliceType           = errors.New("not slice type")
	ErrNotSliceOrPointer      = errors.New("not a slice or pointer to slice")
	ErrElemTypeNotMatchSlice  = errors.New("elem type does not match slice")
	ErrElemNotStructOrPointer = errors.New("elem is not struct or pointer to struct")
	ErrStructFieldNotExists   = errors.New("struct filed not exists")
	ErrStructFieldIsNotInt    = errors.New("struct field is not integer or pointer")
	ErrStructFieldIsNotStr    = errors.New("struct field is not string or pointer")
)

var int64Type = reflect.TypeOf(int64(0))
var stringType = reflect.TypeOf("")

func InSlice(slice interface{}, elem interface{}) bool {
	if slice == nil || elem == nil {
		return false
	}
	sliceTyp := reflect.TypeOf(slice)
	elemTyp := reflect.TypeOf(elem)
	if sliceTyp.Kind() == reflect.Slice &&
		sliceTyp.Elem().Kind() == elemTyp.Kind() {
		switch elemTyp.Kind() {
		case reflect.Int64, reflect.Uint64:
			return InInt64s(ToInt64s_(slice), _int64(elem))
		case reflect.Int, reflect.Uint, reflect.Uintptr:
			if platform64bit {
				return InInt64s(ToInt64s_(slice), _int64(elem))
			}
		case reflect.String:
			return InStrings(_Strings(slice), _string(elem))
		}
	}

	sliceVal, intTypeNotMatch := assertSliceAndElemType(reflect.ValueOf(slice), elemTyp)

	if intTypeNotMatch {
		_elemInt := reflectInt(reflect.ValueOf(elem))
		for i := 0; i < sliceVal.Len(); i++ {
			_sliceInt := reflectInt(sliceVal.Index(i))
			if _elemInt == _sliceInt {
				return true
			}
		}
		return false
	}

	for i := 0; i < sliceVal.Len(); i++ {
		if elem == sliceVal.Index(i).Interface() {
			return true
		}
	}
	return false
}

func InInt64s(slice []int64, elem int64) bool {
	for _, x := range slice {
		if x == elem {
			return true
		}
	}
	return false
}

func InStrings(slice []string, elem string) bool {
	for _, x := range slice {
		if elem == x {
			return true
		}
	}
	return false
}

func InsertSlice(slice interface{}, index int, elem interface{}) (out interface{}) {
	if slice == nil || elem == nil {
		return slice
	}
	sliceTyp := reflect.TypeOf(slice)
	elemTyp := reflect.TypeOf(elem)
	if sliceTyp.Kind() == reflect.Slice &&
		sliceTyp.Elem().Kind() == elemTyp.Kind() {
		switch elemTyp.Kind() {
		case reflect.Int64, reflect.Uint64:
			return InsertInt64s(ToInt64s_(slice), index, _int64(elem))
		case reflect.Int, reflect.Uint, reflect.Uintptr:
			if platform64bit {
				return InsertInt64s(ToInt64s_(slice), index, _int64(elem))
			}
		case reflect.String:
			return InsertStrings(_Strings(slice), index, _string(elem))
		}
	}

	sliceVal, intTypeNotMatch := assertSliceAndElemType(reflect.ValueOf(slice), elemTyp)

	outVal := reflect.MakeSlice(sliceVal.Type(), 0, sliceVal.Len()+1)
	for i := 0; i < sliceVal.Len() && i < index; i++ {
		outVal = reflect.Append(outVal, sliceVal.Index(i))
	}
	if intTypeNotMatch {
		_elemInt := reflectInt(reflect.ValueOf(elem))
		_sliceInt := reflect.New(sliceTyp.Elem())
		_sliceInt.Elem().SetInt(_elemInt)
		outVal = reflect.Append(outVal, reflect.Indirect(_sliceInt))
	} else {
		outVal = reflect.Append(outVal, reflect.ValueOf(elem))
	}
	for i := index; i < sliceVal.Len(); i++ {
		outVal = reflect.Append(outVal, sliceVal.Index(i))
	}
	return outVal.Interface()
}

func InsertInt64s(slice []int64, index int, elem int64) (out Int64s) {
	out = make([]int64, 0, len(slice)+1)
	if len(slice) < index {
		out = append(out, slice...)
		out = append(out, elem)
		return
	}
	out = append(out, slice[:index]...)
	out = append(out, elem)
	out = append(out, slice[index:]...)
	return
}

func InsertStrings(slice []string, index int, elem string) (out Strings) {
	out = make([]string, 0, len(slice)+1)
	if len(slice) < index {
		out = append(out, slice...)
		out = append(out, elem)
		return
	}
	out = append(out, slice[:index]...)
	out = append(out, elem)
	out = append(out, slice[index:]...)
	return
}

func ReverseSlice(slice interface{}) interface{} {
	if slice == nil {
		return slice
	}
	switch slice := slice.(type) {
	case Int64s, []int64, []uint64:
		return ReverseInt64s(ToInt64s_(slice))
	case Strings, []string:
		return ReverseStrings(_Strings(slice))
	}

	sliceVal := reflect.ValueOf(slice)
	if sliceVal.Kind() != reflect.Slice {
		panic(ErrNotSliceType)
	}
	outVal := reflect.MakeSlice(sliceVal.Type(), 0, sliceVal.Len())
	for i := sliceVal.Len() - 1; i >= 0; i-- {
		outVal = reflect.Append(outVal, sliceVal.Index(i))
	}
	return outVal.Interface()
}

func ReverseInt64s(slice []int64) Int64s {
	out := make([]int64, 0, len(slice))
	for i := len(slice) - 1; i >= 0; i-- {
		out = append(out, slice[i])
	}
	return out
}

func ReverseStrings(slice []string) Strings {
	out := make([]string, 0, len(slice))
	for i := len(slice) - 1; i >= 0; i-- {
		out = append(out, slice[i])
	}
	return out
}

func DiffInt64s(a []int64, b []int64) Int64s {
	bset := make(map[int64]struct{}, len(b))
	for _, x := range b {
		bset[x] = struct{}{}
	}
	out := make([]int64, 0)
	for _, x := range a {
		if _, ok := bset[x]; !ok {
			out = append(out, x)
		}
	}
	return out
}

func DiffStrings(a []string, b []string) Strings {
	bset := make(map[string]struct{}, len(b))
	for _, x := range b {
		bset[x] = struct{}{}
	}
	out := make([]string, 0)
	for _, x := range a {
		if _, ok := bset[x]; !ok {
			out = append(out, x)
		}
	}
	return out
}

func Pluck(slice interface{}, field string) interface{} {
	if slice == nil {
		return slice
	}
	sliceVal := reflect.ValueOf(slice)
	for sliceVal.Kind() == reflect.Ptr {
		sliceVal = reflect.Indirect(sliceVal)
	}
	sliceTyp := sliceVal.Type()
	fieldInfo := assertSliceElemStructAndField(sliceTyp, field)

	var outVal = reflect.MakeSlice(reflect.SliceOf(fieldInfo.Type), 0, sliceVal.Len())
	for i := 0; i < sliceVal.Len(); i++ {
		elem := reflect.Indirect(sliceVal.Index(i))
		fieldVal := elem.FieldByName(field)
		outVal = reflect.Append(outVal, fieldVal)
	}
	return outVal.Interface()
}

func PluckInt64s(slice interface{}, field string) Int64s {
	if slice == nil {
		return nil
	}
	sliceVal := reflect.ValueOf(slice)
	for sliceVal.Kind() == reflect.Ptr {
		sliceVal = reflect.Indirect(sliceVal)
	}
	sliceTyp := sliceVal.Type()
	fieldInfo := assertSliceElemStructAndField(sliceTyp, field)
	if !(isIntType(fieldInfo.Type) ||
		(fieldInfo.Type.Kind() == reflect.Ptr && isIntType(fieldInfo.Type.Elem()))) {
		panic(ErrStructFieldIsNotInt)
	}

	out := make([]int64, 0, sliceVal.Len())
	for i := 0; i < sliceVal.Len(); i++ {
		elem := reflect.Indirect(sliceVal.Index(i))
		fieldVal := reflect.Indirect(elem.FieldByName(field))
		if fieldVal.IsValid() {
			out = append(out, reflectInt(fieldVal))
		}
	}
	return out
}

func PluckStrings(slice interface{}, field string) Strings {
	if slice == nil {
		return nil
	}
	sliceVal := reflect.ValueOf(slice)
	for sliceVal.Kind() == reflect.Ptr {
		sliceVal = reflect.Indirect(sliceVal)
	}
	sliceTyp := sliceVal.Type()
	fieldInfo := assertSliceElemStructAndField(sliceTyp, field)
	if !(fieldInfo.Type.Kind() == reflect.String ||
		(fieldInfo.Type.Kind() == reflect.Ptr && fieldInfo.Type.Elem().Kind() == reflect.String)) {
		panic(ErrStructFieldIsNotStr)
	}

	out := make([]string, 0, sliceVal.Len())
	for i := 0; i < sliceVal.Len(); i++ {
		elem := reflect.Indirect(sliceVal.Index(i))
		fieldVal := reflect.Indirect(elem.FieldByName(field))
		if fieldVal.IsValid() {
			out = append(out, fieldVal.String())
		}
	}
	return out
}

func ToMap(slice interface{}, keyField string) interface{} {
	if slice == nil {
		return nil
	}
	sliceVal := reflect.ValueOf(slice)
	for sliceVal.Kind() == reflect.Ptr {
		sliceVal = reflect.Indirect(sliceVal)
	}
	sliceTyp := sliceVal.Type()
	elemTyp := sliceTyp.Elem()
	fieldInfo := assertSliceElemStructAndField(sliceTyp, keyField)
	keyTyp := fieldInfo.Type
	if keyTyp.Kind() == reflect.Ptr {
		keyTyp = keyTyp.Elem()
	}

	outVal := reflect.MakeMapWithSize(reflect.MapOf(keyTyp, elemTyp), sliceVal.Len())
	for i := 0; i < sliceVal.Len(); i++ {
		elem := sliceVal.Index(i)
		fieldVal := reflect.Indirect(reflect.Indirect(elem).FieldByName(keyField))
		outVal.SetMapIndex(fieldVal, elem)
	}
	return outVal.Interface()
}

func ToInt64Map(slice interface{}, keyField string) interface{} {
	if slice == nil {
		return nil
	}
	sliceVal := reflect.ValueOf(slice)
	for sliceVal.Kind() == reflect.Ptr {
		sliceVal = reflect.Indirect(sliceVal)
	}
	sliceTyp := sliceVal.Type()
	elemTyp := sliceTyp.Elem()
	fieldInfo := assertSliceElemStructAndField(sliceTyp, keyField)
	if !(isIntType(fieldInfo.Type) ||
		(fieldInfo.Type.Kind() == reflect.Ptr && isIntType(fieldInfo.Type.Elem()))) {
		panic(ErrStructFieldIsNotInt)
	}

	outVal := reflect.MakeMapWithSize(reflect.MapOf(int64Type, elemTyp), sliceVal.Len())
	for i := 0; i < sliceVal.Len(); i++ {
		elem := sliceVal.Index(i)
		fieldVal := reflect.Indirect(reflect.Indirect(elem).FieldByName(keyField))
		key := reflectInt(fieldVal)
		outVal.SetMapIndex(reflect.ValueOf(key), elem)
	}
	return outVal.Interface()
}

func ToStringMap(slice interface{}, keyField string) interface{} {
	if slice == nil {
		return nil
	}
	sliceVal := reflect.ValueOf(slice)
	for sliceVal.Kind() == reflect.Ptr {
		sliceVal = reflect.Indirect(sliceVal)
	}
	sliceTyp := sliceVal.Type()
	elemTyp := sliceTyp.Elem()
	fieldInfo := assertSliceElemStructAndField(sliceTyp, keyField)
	if !(fieldInfo.Type.Kind() == reflect.String ||
		(fieldInfo.Type.Kind() == reflect.Ptr && fieldInfo.Type.Elem().Kind() == reflect.String)) {
		panic(ErrStructFieldIsNotStr)
	}

	outVal := reflect.MakeMapWithSize(reflect.MapOf(stringType, elemTyp), sliceVal.Len())
	for i := 0; i < sliceVal.Len(); i++ {
		elem := sliceVal.Index(i)
		fieldVal := reflect.Indirect(reflect.Indirect(elem).FieldByName(keyField))
		outVal.SetMapIndex(fieldVal, elem)
	}
	return outVal.Interface()
}

func assertSliceAndElemType(sliceVal reflect.Value, elemTyp reflect.Type) (reflect.Value, bool) {
	for sliceVal.Kind() == reflect.Ptr {
		sliceVal = reflect.Indirect(sliceVal)
	}
	if sliceVal.Kind() != reflect.Slice {
		panic(ErrNotSliceOrPointer)
	}
	intTypeNotMatch := false
	sliceTyp := sliceVal.Type()
	if elemTyp != sliceTyp.Elem() {
		// int-family
		if isIntType(sliceTyp.Elem()) && isIntType(elemTyp) {
			intTypeNotMatch = true
		} else {
			panic(ErrElemTypeNotMatchSlice)
		}
	}
	return sliceVal, intTypeNotMatch
}

func assertSliceElemStructAndField(sliceTyp reflect.Type, field string) reflect.StructField {
	if sliceTyp.Kind() != reflect.Slice {
		panic(ErrNotSliceOrPointer)
	}
	elemTyp := sliceTyp.Elem()
	elemIsPtr := elemTyp.Kind() == reflect.Ptr
	if !(elemTyp.Kind() == reflect.Struct ||
		(elemIsPtr && elemTyp.Elem().Kind() == reflect.Struct)) {
		panic(ErrElemNotStructOrPointer)
	}
	var fieldInfo reflect.StructField
	var ok bool
	if elemIsPtr {
		fieldInfo, ok = elemTyp.Elem().FieldByName(field)
	} else {
		fieldInfo, ok = elemTyp.FieldByName(field)
	}
	if !ok {
		panic(ErrStructFieldNotExists)
	}
	return fieldInfo
}

func ParseCommaInts(values string, ignoreZero bool) (slice Int64s, isMalformed bool) {
	values = strings.ReplaceAll(values, " ", "")
	values = strings.Trim(values, ",")
	for _, x := range strings.Split(values, ",") {
		id, err := strconv.ParseInt(x, 10, 64)
		if err != nil {
			isMalformed = true
			continue
		}
		if id == 0 && ignoreZero {
			continue
		}
		slice = append(slice, id)
	}
	return
}
