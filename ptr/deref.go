package ptr

import "time"

func DerefBool(p *bool) bool {
	if p != nil {
		return *p
	}
	return false
}

func DerefString(p *string) string {
	if p != nil {
		return *p
	}
	return ""
}

func DerefInt(p *int) int {
	if p != nil {
		return *p
	}
	return 0
}

func DerefInt8(p *int8) int8 {
	if p != nil {
		return *p
	}
	return 0
}

func DerefInt16(p *int16) int16 {
	if p != nil {
		return *p
	}
	return 0
}

func DerefInt32(p *int32) int32 {
	if p != nil {
		return *p
	}
	return 0
}

func DerefInt64(p *int64) int64 {
	if p != nil {
		return *p
	}
	return 0
}

func DerefUint(p *uint) uint {
	if p != nil {
		return *p
	}
	return 0
}

func DerefUint8(p *uint8) uint8 {
	if p != nil {
		return *p
	}
	return 0
}

func DerefUint16(p *uint16) uint16 {
	if p != nil {
		return *p
	}
	return 0
}

func DerefUint32(p *uint32) uint32 {
	if p != nil {
		return *p
	}
	return 0
}

func DerefUint64(p *uint64) uint64 {
	if p != nil {
		return *p
	}
	return 0
}

func DerefFloat32(p *float32) float32 {
	if p != nil {
		return *p
	}
	return 0
}

func DerefFloat64(p *float64) float64 {
	if p != nil {
		return *p
	}
	return 0
}

func DerefTime(p *time.Time) time.Time {
	if p != nil {
		return *p
	}
	return time.Time{}
}
