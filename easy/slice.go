package easy

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

const (
	errNotSliceType           = "not slice type"
	errNotSliceOrPointer      = "not a slice or pointer to slice"
	errElemTypeNotMatchSlice  = "elem type does not match slice"
	errElemNotStructOrPointer = "elem is not struct or pointer to struct"
	errStructFieldNotProvided = "struct field is not provided"
	errStructFieldNotExists   = "struct field not exists"
	errStructFieldIsNotInt    = "struct field is not integer or pointer"
	errStructFieldIsNotStr    = "struct field is not string or pointer"
	errPredicateFuncSig       = "predicate func signature not match"
)

func panicNilParams(where string, params ...interface{}) {
	const (
		isNilInterface = "%s: param %s is nil interface"
	)
	for i := 0; i < len(params); i += 2 {
		arg := params[i].(string)
		val := params[i+1]
		if val == nil {
			panic(fmt.Sprintf(isNilInterface, where, arg))
		}
	}
}

var int64Type = reflect.TypeOf(int64(0))
var stringType = reflect.TypeOf("")

func InSlice(slice interface{}, elem interface{}) bool {
	if slice == nil {
		return false
	}
	if elem == nil {
		sliceVal := indirect(reflect.ValueOf(slice))
		sliceTyp := sliceVal.Type()
		if sliceTyp.Kind() == reflect.Slice && isNillableKind(sliceTyp.Elem().Kind()) {
			for i := 0; i < sliceVal.Len(); i++ {
				if sliceVal.Index(i).IsNil() {
					return true
				}
			}
		}
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

	sliceVal, intTypeNotMatch := assertSliceAndElemType("InSlice", reflect.ValueOf(slice), elemTyp)

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

func Index(slice interface{}, elem interface{}) int {
	if slice == nil {
		return -1
	}
	if elem == nil {
		sliceVal := indirect(reflect.ValueOf(slice))
		sliceTyp := sliceVal.Type()
		if sliceTyp.Kind() == reflect.Slice && isNillableKind(sliceTyp.Elem().Kind()) {
			for i := 0; i < sliceVal.Len(); i++ {
				if sliceVal.Index(i).IsNil() {
					return i
				}
			}
		}
		return -1
	}

	sliceTyp := reflect.TypeOf(slice)
	elemTyp := reflect.TypeOf(elem)
	if sliceTyp.Kind() == reflect.Slice &&
		sliceTyp.Elem().Kind() == elemTyp.Kind() {
		if _is64bitInt(elemTyp) {
			return IndexInt64s(ToInt64s_(slice), _int64(elem))
		}
		if elemTyp.Kind() == reflect.String {
			return IndexStrings(_Strings(slice), _string(elem))
		}
	}

	sliceVal, intTypeNotMatch := assertSliceAndElemType("Index", reflect.ValueOf(slice), elemTyp)

	if intTypeNotMatch {
		_elemInt := reflectInt(reflect.ValueOf(elem))
		for i := 0; i < sliceVal.Len(); i++ {
			_sliceInt := reflectInt(sliceVal.Index(i))
			if _elemInt == _sliceInt {
				return i
			}
		}
		return -1
	}

	for i := 0; i < sliceVal.Len(); i++ {
		if elem == sliceVal.Index(i).Interface() {
			return i
		}
	}
	return -1
}

func IndexInt64s(slice []int64, elem int64) int {
	for i := 0; i < len(slice); i++ {
		if elem == slice[i] {
			return i
		}
	}
	return -1
}

func IndexStrings(slice []string, elem string) int {
	for i := 0; i < len(slice); i++ {
		if elem == slice[i] {
			return i
		}
	}
	return -1
}

func LastIndex(slice interface{}, elem interface{}) int {
	if slice == nil {
		return -1
	}
	if elem == nil {
		sliceVal := indirect(reflect.ValueOf(slice))
		sliceTyp := sliceVal.Type()
		if sliceTyp.Kind() == reflect.Slice && isNillableKind(sliceTyp.Elem().Kind()) {
			for i := sliceVal.Len() - 1; i >= 0; i-- {
				if sliceVal.Index(i).IsNil() {
					return i
				}
			}
		}
		return -1
	}

	sliceTyp := reflect.TypeOf(slice)
	elemTyp := reflect.TypeOf(elem)
	if sliceTyp.Kind() == reflect.Slice &&
		sliceTyp.Elem().Kind() == elemTyp.Kind() {
		if _is64bitInt(elemTyp) {
			return LastIndexInt64s(ToInt64s_(slice), _int64(elem))
		}
		if elemTyp.Kind() == reflect.String {
			return LastIndexStrings(_Strings(slice), _string(elem))
		}
	}

	sliceVal, intTypeNotMatch := assertSliceAndElemType("LastIndex", reflect.ValueOf(slice), elemTyp)

	if intTypeNotMatch {
		_elemInt := reflectInt(reflect.ValueOf(elem))
		for i := sliceVal.Len() - 1; i >= 0; i-- {
			_sliceInt := reflectInt(sliceVal.Index(i))
			if _elemInt == _sliceInt {
				return i
			}
		}
		return -1
	}

	for i := sliceVal.Len() - 1; i >= 0; i-- {
		if elem == sliceVal.Index(i).Interface() {
			return i
		}
	}
	return -1
}

func LastIndexInt64s(slice []int64, elem int64) int {
	for i := len(slice) - 1; i >= 0; i-- {
		if elem == slice[i] {
			return i
		}
	}
	return -1
}

func LastIndexStrings(slice []string, elem string) int {
	for i := len(slice) - 1; i >= 0; i-- {
		if elem == slice[i] {
			return i
		}
	}
	return -1
}

func InsertSlice(slice interface{}, index int, elem interface{}) (out interface{}) {
	if slice == nil || elem == nil {
		panicNilParams("InsertSlice", "slice", slice, "elem", elem)
	}
	sliceTyp := reflect.TypeOf(slice)
	elemTyp := reflect.TypeOf(elem)
	if sliceTyp.Kind() == reflect.Slice &&
		sliceTyp.Elem().Kind() == elemTyp.Kind() {
		if _is64bitInt(elemTyp) {
			return InsertInt64s(ToInt64s_(slice), index, _int64(elem))
		}
		if elemTyp.Kind() == reflect.String {
			return InsertStrings(_Strings(slice), index, _string(elem))
		}
	}

	sliceVal, intTypeNotMatch := assertSliceAndElemType("InsertSlice", reflect.ValueOf(slice), elemTyp)

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
		panicNilParams("ReverseSlice", "slice", slice)
	}
	switch slice := slice.(type) {
	case Int64s, []int64, []uint64:
		return ReverseInt64s(ToInt64s_(slice))
	case Strings, []string:
		return ReverseStrings(_Strings(slice))
	}

	sliceVal := reflect.ValueOf(slice)
	if sliceVal.Kind() != reflect.Slice {
		panic("ReverseSlice: " + errNotSliceType)
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
		panicNilParams("Pluck", "slice", slice)
	}
	sliceVal := reflect.ValueOf(slice)
	for sliceVal.Kind() == reflect.Ptr {
		sliceVal = reflect.Indirect(sliceVal)
	}
	sliceTyp := sliceVal.Type()
	fieldInfo := assertSliceElemStructAndField("Pluck", sliceTyp, field)

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
		panicNilParams("PluckInt64s", "slice", slice)
	}
	sliceVal := reflect.ValueOf(slice)
	for sliceVal.Kind() == reflect.Ptr {
		sliceVal = reflect.Indirect(sliceVal)
	}
	sliceTyp := sliceVal.Type()
	fieldInfo := assertSliceElemStructAndField("PluckInt64s", sliceTyp, field)
	if !(isIntType(fieldInfo.Type) ||
		(fieldInfo.Type.Kind() == reflect.Ptr && isIntType(fieldInfo.Type.Elem()))) {
		panic("PluckInt64s: " + errStructFieldIsNotInt)
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
		panicNilParams("PluckStrings", "slice", slice)
	}
	sliceVal := reflect.ValueOf(slice)
	for sliceVal.Kind() == reflect.Ptr {
		sliceVal = reflect.Indirect(sliceVal)
	}
	sliceTyp := sliceVal.Type()
	fieldInfo := assertSliceElemStructAndField("PluckStrings", sliceTyp, field)
	if !(fieldInfo.Type.Kind() == reflect.String ||
		(fieldInfo.Type.Kind() == reflect.Ptr && fieldInfo.Type.Elem().Kind() == reflect.String)) {
		panic("PluckStrings: " + errStructFieldIsNotStr)
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
		panicNilParams("ToMap", "slice", slice)
	}
	sliceVal := reflect.ValueOf(slice)
	for sliceVal.Kind() == reflect.Ptr {
		sliceVal = reflect.Indirect(sliceVal)
	}
	sliceTyp := sliceVal.Type()
	elemTyp := sliceTyp.Elem()
	fieldInfo := assertSliceElemStructAndField("ToMap", sliceTyp, keyField)
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
		panicNilParams("ToInt64Map", "slice", slice)
	}
	sliceVal := reflect.ValueOf(slice)
	for sliceVal.Kind() == reflect.Ptr {
		sliceVal = reflect.Indirect(sliceVal)
	}
	sliceTyp := sliceVal.Type()
	elemTyp := sliceTyp.Elem()
	fieldInfo := assertSliceElemStructAndField("ToInt64Map", sliceTyp, keyField)
	if !(isIntType(fieldInfo.Type) ||
		(fieldInfo.Type.Kind() == reflect.Ptr && isIntType(fieldInfo.Type.Elem()))) {
		panic("ToInt64Map: " + errStructFieldIsNotInt)
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
		panicNilParams("ToStringMap", "slice", slice)
	}
	sliceVal := reflect.ValueOf(slice)
	for sliceVal.Kind() == reflect.Ptr {
		sliceVal = reflect.Indirect(sliceVal)
	}
	sliceTyp := sliceVal.Type()
	elemTyp := sliceTyp.Elem()
	fieldInfo := assertSliceElemStructAndField("ToStringMap", sliceTyp, keyField)
	if !(fieldInfo.Type.Kind() == reflect.String ||
		(fieldInfo.Type.Kind() == reflect.Ptr && fieldInfo.Type.Elem().Kind() == reflect.String)) {
		panic("ToStringMap: " + errStructFieldIsNotStr)
	}

	outVal := reflect.MakeMapWithSize(reflect.MapOf(stringType, elemTyp), sliceVal.Len())
	for i := 0; i < sliceVal.Len(); i++ {
		elem := sliceVal.Index(i)
		fieldVal := reflect.Indirect(reflect.Indirect(elem).FieldByName(keyField))
		outVal.SetMapIndex(fieldVal, elem)
	}
	return outVal.Interface()
}

func Find(slice interface{}, predicate interface{}) interface{} {
	if slice == nil || predicate == nil {
		panicNilParams("Find", "slice", slice, "predicate", predicate)
	}
	sliceVal := reflect.ValueOf(slice)
	fnVal := reflect.ValueOf(predicate)
	sliceVal = assertSliceAndPredicateFunc("Find", sliceVal, fnVal.Type())

	outVal := reflect.New(sliceVal.Type().Elem())
	for i := 0; i < sliceVal.Len(); i++ {
		elem := sliceVal.Index(i)
		match := fnVal.Call([]reflect.Value{elem})[0].Interface().(bool)
		if match {
			outVal.Elem().Set(elem)
			break
		}
	}
	return outVal.Elem().Interface()
}

func Filter(slice interface{}, predicate interface{}) interface{} {
	if slice == nil || predicate == nil {
		panicNilParams("Filter", "slice", slice, "predicate", predicate)
	}
	sliceVal := reflect.ValueOf(slice)
	fnVal := reflect.ValueOf(predicate)
	sliceVal = assertSliceAndPredicateFunc("Filter", sliceVal, fnVal.Type())

	outVal := reflect.MakeSlice(sliceVal.Type(), 0, 1)
	for i := 0; i < sliceVal.Len(); i++ {
		elem := sliceVal.Index(i)
		match := fnVal.Call([]reflect.Value{elem})[0].Interface().(bool)
		if match {
			outVal = reflect.Append(outVal, elem)
		}
	}
	return outVal.Interface()
}

func assertSliceAndElemType(where string, sliceVal reflect.Value, elemTyp reflect.Type) (reflect.Value, bool) {
	sliceVal = indirect(sliceVal)
	if sliceVal.Kind() != reflect.Slice {
		panic(where + ": " + errNotSliceOrPointer)
	}
	intTypeNotMatch := false
	sliceTyp := sliceVal.Type()
	if elemTyp != sliceTyp.Elem() {
		// int-family
		if isIntType(sliceTyp.Elem()) && isIntType(elemTyp) {
			intTypeNotMatch = true
		} else {
			panic(where + ": " + errElemTypeNotMatchSlice)
		}
	}
	return sliceVal, intTypeNotMatch
}

func assertSliceElemStructAndField(where string, sliceTyp reflect.Type, field string) reflect.StructField {
	if field == "" {
		panic(where + ": " + errStructFieldNotProvided)
	}
	if sliceTyp.Kind() != reflect.Slice {
		panic(where + ": " + errNotSliceOrPointer)
	}
	elemTyp := sliceTyp.Elem()
	elemIsPtr := elemTyp.Kind() == reflect.Ptr
	if !(elemTyp.Kind() == reflect.Struct ||
		(elemIsPtr && elemTyp.Elem().Kind() == reflect.Struct)) {
		panic(where + ": " + errElemNotStructOrPointer)
	}
	var fieldInfo reflect.StructField
	var ok bool
	if elemIsPtr {
		fieldInfo, ok = elemTyp.Elem().FieldByName(field)
	} else {
		fieldInfo, ok = elemTyp.FieldByName(field)
	}
	if !ok {
		panic(where + ": " + errStructFieldNotExists)
	}
	return fieldInfo
}

func assertSliceAndPredicateFunc(where string, sliceVal reflect.Value, fnTyp reflect.Type) reflect.Value {
	sliceVal = indirect(sliceVal)
	if sliceVal.Kind() != reflect.Slice {
		panic(where + ": " + errNotSliceOrPointer)
	}
	elemTyp := sliceVal.Type().Elem()
	if !(fnTyp.Kind() == reflect.Func &&
		fnTyp.NumIn() == 1 && fnTyp.NumOut() == 1 &&
		(fnTyp.In(0).Kind() == reflect.Interface || fnTyp.In(0) == elemTyp) &&
		fnTyp.Out(0).Kind() == reflect.Bool) {
		panic(where + ": " + errPredicateFuncSig)
	}
	return sliceVal
}

func ParseCommaInt64s(values string, ignoreZero bool) (slice Int64s, isMalformed bool) {
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

type IJ struct{ I, J int }

// SplitBatch
func SplitBatch(total, batch int) []IJ {
	if total <= 0 {
		return nil
	}
	if batch <= 0 {
		return []IJ{{0, total}}
	}
	ret := make([]IJ, 0, total/batch+1)
	for i, j := 0, batch; i < total; i, j = i+batch, j+batch {
		if j > total {
			j = total
		}
		ret = append(ret, IJ{i, j})
	}
	return ret
}

func indirect(value reflect.Value) reflect.Value {
	for value.Kind() == reflect.Ptr {
		value = reflect.Indirect(value)
	}
	return value
}
