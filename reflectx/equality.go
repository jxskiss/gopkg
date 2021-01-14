package reflectx

import (
	"fmt"
	"reflect"
)

func IsIdenticalType(a, b interface{}) (equal bool, msg string) {
	typ1 := reflect.TypeOf(a)
	typ2 := reflect.TypeOf(b)
	return isEqualType(typ1, typ2)
}

func isEqualType(typ1, typ2 reflect.Type) (bool, string) {
	if typ1.Kind() != typ2.Kind() {
		return false, fmt.Sprintf("type kind not equal: %s, %s", _t(typ1), _t(typ2))
	}
	if typ1.Kind() == reflect.Ptr {
		if typ1.Elem().Kind() != typ2.Elem().Kind() {
			return false, fmt.Sprintf("pointer type not equal: %s, %s", _t(typ1), _t(typ2))
		}
		typ1 = typ1.Elem()
		typ2 = typ2.Elem()
		return isEqualType(typ1, typ2)
	}
	switch typ1.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128, reflect.String:
		return true, ""
	case reflect.Struct:
		return isEqualStruct(typ1, typ2)
	case reflect.Slice:
		return isEqualSlice(typ1, typ2)
	case reflect.Map:
		return isEqualMap(typ1, typ2)
	}
	return false, fmt.Sprintf("unsupported types: %s, %s", _t(typ1), _t(typ2))
}

func isEqualStruct(typ1, typ2 reflect.Type) (bool, string) {
	if typ1.Kind() != reflect.Struct || typ1.Kind() != reflect.Struct {
		return false, fmt.Sprintf("type is not struct: %s, %s", _t(typ1), _t(typ2))
	}
	if typ1.NumField() != typ2.NumField() {
		return false, fmt.Sprintf("struct field num not match: %s, %s", _t(typ1), _t(typ2))
	}
	fnum := typ1.NumField()
	for i := 0; i < fnum; i++ {
		f1 := typ1.Field(i)
		f2 := typ2.Field(i)
		if equal, msg := isEqualField(typ1, f1, f2); !equal {
			return false, fmt.Sprintf("struct field not equal: %s", msg)
		}
	}
	return true, ""
}

func isEqualField(typ reflect.Type, field1, field2 reflect.StructField) (bool, string) {
	if field1.Name != field2.Name {
		return false, fmt.Sprintf("field name not equal: %s, %s", _f(typ, field1), _f(typ, field2))
	}
	if field1.Offset != field2.Offset {
		return false, fmt.Sprintf("field offset not equal: %s, %s", _f(typ, field1), _f(typ, field2))
	}

	typ1 := field1.Type
	typ2 := field2.Type
	if typ1.Kind() != typ2.Kind() {
		return false, fmt.Sprintf("field type not equal: %s, %s", _f(typ, field1), _f(typ, field2))
	}
	if typ1.Kind() == reflect.Ptr {
		typ1 = typ1.Elem()
		typ2 = typ2.Elem()
	}
	equal, msg := isEqualType(typ1, typ2)
	if equal {
		return true, ""
	}
	return false, fmt.Sprintf("field type not euqal: %s", msg)
}

func isEqualSlice(typ1, typ2 reflect.Type) (bool, string) {
	if typ1.Kind() != reflect.Slice || typ1.Kind() != reflect.Slice {
		return false, fmt.Sprintf("type is not slice: %s, %s", _t(typ1), _t(typ2))
	}
	typ1 = typ1.Elem()
	typ2 = typ2.Elem()
	return isEqualType(typ1, typ2)
}

func isEqualMap(typ1, typ2 reflect.Type) (bool, string) {
	if typ1.Kind() != reflect.Map || typ2.Kind() != reflect.Map {
		return false, fmt.Sprintf("type is not map: %s, %s", _t(typ1), _t(typ2))
	}

	keyTyp1 := typ1.Key()
	keyTyp2 := typ2.Key()
	if equal, msg := isEqualType(keyTyp1, keyTyp2); !equal {
		return false, msg
	}

	elemTyp1 := typ1.Elem()
	elemTyp2 := typ2.Elem()
	if equal, msg := isEqualType(elemTyp1, elemTyp2); !equal {
		return false, msg
	}

	return true, ""
}

func _t(typ reflect.Type) string {
	return fmt.Sprintf("%s.%s", typ.PkgPath(), typ.Name())
}

func _f(typ reflect.Type, field reflect.StructField) string {
	return fmt.Sprintf("%s.%s", _t(typ), field.Name)
}
