package easy

import (
	"github.com/jxskiss/gopkg/reflectx"
	"reflect"
)

func _int32(x interface{}) int32       { return int32(reflectx.CastInt(x)) }
func _int64(x interface{}) int64       { return reflectx.CastInt(x) }
func _string(x interface{}) string     { return reflectx.CastString(x) }
func reflectInt(v reflect.Value) int64 { return reflectx.ReflectInt(v) }
