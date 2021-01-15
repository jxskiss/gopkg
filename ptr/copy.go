package ptr

import (
	"reflect"
	"time"
)

func CopyBool(p *bool) *bool {
	if p == nil {
		return nil
	}
	r := *p
	return &r
}

func CopyString(p *string) *string {
	if p == nil {
		return nil
	}
	r := *p
	return &r
}

func CopyInt(p *int) *int {
	if p == nil {
		return nil
	}
	r := *p
	return &r
}

func CopyInt8(p *int8) *int8 {
	if p == nil {
		return nil
	}
	r := *p
	return &r
}

func CopyInt16(p *int16) *int16 {
	if p == nil {
		return nil
	}
	r := *p
	return &r
}

func CopyInt32(p *int32) *int32 {
	if p == nil {
		return nil
	}
	r := *p
	return &r
}

func CopyInt64(p *int64) *int64 {
	if p == nil {
		return nil
	}
	r := *p
	return &r
}

func CopyUint(p *uint) *uint {
	if p == nil {
		return nil
	}
	r := *p
	return &r
}

func CopyUint8(p *uint8) *uint8 {
	if p == nil {
		return nil
	}
	r := *p
	return &r
}

func CopyUint16(p *uint16) *uint16 {
	if p == nil {
		return nil
	}
	r := *p
	return &r
}

func CopyUint32(p *uint32) *uint32 {
	if p == nil {
		return nil
	}
	r := *p
	return &r
}

func CopyUint64(p *uint64) *uint64 {
	if p == nil {
		return nil
	}
	r := *p
	return &r
}

func CopyFloat32(p *float32) *float32 {
	if p == nil {
		return nil
	}
	r := *p
	return &r
}

func CopyFloat64(p *float64) *float64 {
	if p == nil {
		return nil
	}
	r := *p
	return &r
}

func CopyTime(p *time.Time) *time.Time {
	if p == nil {
		return nil
	}
	r := *p
	return &r
}

func CopyDuration(p *time.Duration) *time.Duration {
	if p == nil {
		return nil
	}
	r := *p
	return &r
}

// CopyAny return a shallow copy of the given pointer of any type.
func CopyAny(v interface{}) interface{} {
	val := reflect.Indirect(reflect.ValueOf(v))
	p := reflect.New(val.Type())
	p.Elem().Set(val)
	return p.Interface()
}
