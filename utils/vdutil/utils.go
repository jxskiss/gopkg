package validat

import (
	"fmt"
	"reflect"
	"strconv"
	"unsafe"
)

type Int64OrString interface {
	int64 | string
}

func parseInt64[T Int64OrString](value T) (int64, error) {
	switch reflect.TypeOf((*T)(nil)).Elem().Kind() {
	case reflect.Int64:
		return any(value).(int64), nil
	case reflect.String:
		s := any(value).(string)
		intVal, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("value %s is not integer", s)
		}
		return intVal, nil
	}
	panic("bug: unreachable code")
}

func parseInt64s[T Int64OrString](slice []T) ([]int64, error) {
	switch reflect.TypeOf((*T)(nil)).Elem().Kind() {
	case reflect.Int64:
		return *(*[]int64)(unsafe.Pointer(&slice)), nil
	case reflect.String:
		slice := *(*[]string)(unsafe.Pointer(&slice))
		out := make([]int64, 0, len(slice))
		for _, s := range slice {
			intVal, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("value %s is not interger", s)
			}
			out = append(out, intVal)
		}
		return out, nil
	}
	panic("bug: unreachable code")
}
