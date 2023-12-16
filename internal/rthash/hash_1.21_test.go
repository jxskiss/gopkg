//go:build gc && go1.21

package rthash

import (
	"fmt"
	"testing"
)

func TestHashFunc_Interface(t *testing.T) {
	t.Run("efaceHash / empty interface", func(t *testing.T) {
		type myEface any
		f := NewHashFunc[myEface]()
		for _, x := range []myEface{
			int8(123),
			"a string",
			complex(1, 2),
			hashable{12345, "abcde"},
			&hashable{12345, "abcde"},
		} {
			hash := f(x)
			t.Logf("%T: %v, hash: %d", x, x, hash)
		}
	})

	t.Run("efaceHash / non-empty interface", func(t *testing.T) {
		f := NewHashFunc[fmt.Stringer]()
		for _, x := range []fmt.Stringer{
			stringerInt(12345),
			stringerStr("abc"),
			hashable{12345, "abcde"},
			&hashable{12345, "abcde"},
		} {
			hash := f(x)
			t.Logf("%T: %v, hash: %d", x, x, hash)
		}
	})
}

type stringerInt int

func (x stringerInt) String() string {
	return fmt.Sprint(int(x))
}

type stringerStr string

func (x stringerStr) String() string {
	return string(x)
}

func (x hashable) String() string {
	return fmt.Sprintf("%d %s", x.A, x.B)
}

func BenchmarkHashFunc_efaceHash(b *testing.B) {
	f := NewHashFunc[any]()
	for i := 0; i < b.N; i++ {
		_ = f(hashable{12345, "abcde"})
	}
}

func BenchmarkHashFunc_typeHash(b *testing.B) {
	f := NewHashFunc[hashable]()
	for i := 0; i < b.N; i++ {
		_ = f(hashable{12345, "abcde"})
	}
}
