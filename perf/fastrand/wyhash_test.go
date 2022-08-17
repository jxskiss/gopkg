package fastrand

import (
	"reflect"
	"testing"
)

func TestWyrandRead(t *testing.T) {
	const seed = 123456789
	buf := make([]byte, 30)
	_wyread(seed, buf)
	want := []byte{111, 76, 184, 83, 145, 188, 211, 99, 49, 80, 32, 94, 112, 213, 124, 90, 117, 91, 135, 96, 163, 125, 165, 132, 168, 83, 84, 13, 72, 142}
	if !reflect.DeepEqual(buf, want) {
		t.Errorf("got %v != want %v", buf, want)
	}
}
