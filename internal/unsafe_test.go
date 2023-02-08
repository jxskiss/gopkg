package internal

import (
	"testing"
	"unsafe"
)

func TestEfaceOf(t *testing.T) {
	var x any = EmptyInterface{}
	ef := EFaceOf(&x)
	if x != *(*any)(unsafe.Pointer(ef)) {
		t.Fatalf("test EfaceOf got unexpected result")
	}
}

func TestUnpackSlice(t *testing.T) {
	var data any = []int{1, 2, 3}
	sh := UnpackSlice(data)
	var got any = *(*[]int)(unsafe.Pointer(&sh))
	if a, b := EFaceOf(&got).RType, EFaceOf(&data).RType; a != b {
		t.Fatalf("test UnpackSlice got different RType, got= %x, data= %x", a, b)
	}
	if a, b := UnpackSlice(got).Data, UnpackSlice(data).Data; a != b {
		t.Fatalf("test UnpackSlice got different Word, got= %x, data= %x", a, b)
	}
}

func TestCastInt(t *testing.T) {
	var data = []any{
		int8(1),
		uint8(2),
		int16(3),
		uint16(4),
		int32(5),
		uint32(6),
		int64(7),
		uint64(8),
		int(9),
		uint(10),
		uintptr(11),
	}
	for i := 0; i < len(data); i++ {
		var want = int64(i + 1)
		got := CastInt(data[i])
		if got != want {
			t.Fatalf("test CastInt, got (%v) != want (%v)", got, want)
		}
	}
}

func TestCastIntPointer(t *testing.T) {
	// pass
}
