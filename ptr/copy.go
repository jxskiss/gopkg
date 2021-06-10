package ptr

import (
	"reflect"
	"time"
)

// CopyBool returns a copy pointer of value *p.
// It returns nil if p is nil.
func CopyBool(p *bool) *bool {
	if p == nil {
		return nil
	}
	r := *p
	return &r
}

// CopyString returns a copy pointer of value *p.
// It returns nil if p is nil.
func CopyString(p *string) *string {
	if p == nil {
		return nil
	}
	r := *p
	return &r
}

// CopyInt returns a copy pointer of value *p.
// It returns nil if p is nil.
func CopyInt(p *int) *int {
	if p == nil {
		return nil
	}
	r := *p
	return &r
}

// CopyInt8 returns a copy pointer of value *p.
// It returns nil if p is nil.
func CopyInt8(p *int8) *int8 {
	if p == nil {
		return nil
	}
	r := *p
	return &r
}

// CopyInt16 returns a copy pointer of value *p.
// It returns nil if p is nil.
func CopyInt16(p *int16) *int16 {
	if p == nil {
		return nil
	}
	r := *p
	return &r
}

// CopyInt32 returns a copy pointer of value *p.
// It returns nil if p is nil.
func CopyInt32(p *int32) *int32 {
	if p == nil {
		return nil
	}
	r := *p
	return &r
}

// CopyInt64 returns a copy pointer of value *p.
// It returns nil if p is nil.
func CopyInt64(p *int64) *int64 {
	if p == nil {
		return nil
	}
	r := *p
	return &r
}

// CopyUint returns a copy pointer of value *p.
// It returns nil if p is nil.
func CopyUint(p *uint) *uint {
	if p == nil {
		return nil
	}
	r := *p
	return &r
}

// CopyUint8 returns a copy pointer of value *p.
// It returns nil if p is nil.
func CopyUint8(p *uint8) *uint8 {
	if p == nil {
		return nil
	}
	r := *p
	return &r
}

// CopyUint16 returns a copy pointer of value *p.
// It returns nil if p is nil.
func CopyUint16(p *uint16) *uint16 {
	if p == nil {
		return nil
	}
	r := *p
	return &r
}

// CopyUint32 returns a copy pointer of value *p.
// It returns nil if p is nil.
func CopyUint32(p *uint32) *uint32 {
	if p == nil {
		return nil
	}
	r := *p
	return &r
}

// CopyUint64 returns a copy pointer of value *p.
// It returns nil if p is nil.
func CopyUint64(p *uint64) *uint64 {
	if p == nil {
		return nil
	}
	r := *p
	return &r
}

// CopyFloat32 returns a copy pointer of value *p.
// It returns nil if p is nil.
func CopyFloat32(p *float32) *float32 {
	if p == nil {
		return nil
	}
	r := *p
	return &r
}

// CopyFloat64 returns a copy pointer of value *p.
// It returns nil if p is nil.
func CopyFloat64(p *float64) *float64 {
	if p == nil {
		return nil
	}
	r := *p
	return &r
}

// CopyTime returns a copy pointer of value *p.
// It returns nil if p is nil.
func CopyTime(p *time.Time) *time.Time {
	if p == nil {
		return nil
	}
	r := *p
	return &r
}

// CopyDuration returns a copy pointer of value *p.
// It returns nil if p is nil.
func CopyDuration(p *time.Duration) *time.Duration {
	if p == nil {
		return nil
	}
	r := *p
	return &r
}

// CopyAny return a shallow copy of the given pointer or value of any type.
// It always returns a pointer if v is not nil.
// If v is not a pointer or value of simple types, it may panic.
//
// If v is nil, it returns a nil interface{}.
// But note that if v is a nil pointer (not a nil interface{}),
// it returns a nil pointer of the same type.
func CopyAny(v interface{}) interface{} {
	if v == nil {
		return nil
	}

	val := reflect.ValueOf(v)
	typ := val.Type()
	isNil := false
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		isNil = val.IsNil()
	}
	if isNil {
		p := reflect.New(reflect.PtrTo(typ))
		return p.Elem().Interface()
	}
	p := reflect.New(typ)
	p.Elem().Set(reflect.Indirect(val))
	return p.Interface()
}
