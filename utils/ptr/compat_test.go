package ptr

import (
	"fmt"
	"math"
	"testing"
)

func TestString(t *testing.T) {
	type SomeString string

	want := "123"
	got := []*string{
		String("123"),
		String(SomeString("123")),
		String(123),
		String(int32(123)),
		String(int64(123)),
	}
	for _, x := range got {
		if *x != want {
			t.Fatalf("want string %q, got %+v", want, x)
		}
	}

	want2 := fmt.Sprint(uint64(math.MaxUint64))
	got2 := String(uint64(math.MaxUint64))
	if *got2 != want2 {
		t.Fatalf("want string %q, got %+v", want2, got2)
	}
}

func TestInt(t *testing.T) {
	want := 123
	got := []*int{
		Int(int8(123)),
		Int(int16(123)),
		Int(int64(123)),
		Int(uint8(123)),
		Int(uintptr(123)),
	}
	for _, x := range got {
		if *x != want {
			t.Fatalf("want int %d, got %+v", want, x)
		}
	}
}

func TestFloat64(t *testing.T) {
	want := float64(123)
	got := []*float64{
		Float64(int8(123)),
		Float64(uint8(123)),
		Float64(int(123)),
		Float64(float32(123)),
		Float64(float64(123)),
	}
	for _, x := range got {
		if *x != want {
			t.Fatalf("want float64 %f, got %+v", want, x)
		}
	}
}

func TestDerefString(t *testing.T) {
	type SomeString string

	testCases := []*struct {
		want string
		got  string
	}{
		{"", DerefString((*string)(nil))},
		{"1234", DerefString(String("1234"))},
		{"", DerefString((*SomeString)(nil))},
		{"1234", DerefString(Ptr(SomeString("1234")))},
	}
	for _, tc := range testCases {
		if tc.got != tc.want {
			t.Fatalf("want string %q, got %+v", tc.want, tc.got)
		}
	}
}

func TestDerefFloat32(t *testing.T) {
	var x *int
	got1 := DerefFloat32(x)
	if got1 != float32(0) {
		t.Fatalf("want float32 zero, but got %+v", got1)
	}

	want := float32(123)
	got2 := []float32{
		DerefFloat32(Ptr(int8(123))),
		DerefFloat32(Ptr(uint16(123))),
		DerefFloat32(Ptr(float32(123))),
		DerefFloat32(Ptr(float64(123))),
	}
	for _, x := range got2 {
		if x != want {
			t.Fatalf("want float32 %f, got %+v", want, x)
		}
	}
}
