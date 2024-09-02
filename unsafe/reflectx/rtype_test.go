package reflectx

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"

	"github.com/jxskiss/gopkg/v2/internal/linkname"
	"github.com/jxskiss/gopkg/v2/internal/unsafeheader"
)

type simple struct {
	A string
}

func TestRTypeMethods(t *testing.T) {
	reflectTyp := reflect.TypeOf((*reflect.Type)(nil)).Elem()
	rtyp := reflect.TypeOf((*RType)(nil))

checkMethod:
	for i := 0; i < reflectTyp.NumMethod(); i++ {
		meth := reflectTyp.Method(i)
		for _, x := range []string{
			// Private methods.
			"common", "uncommon",
			// Unsupported new methods since Go 1.23.
			"OverflowComplex", "OverflowFloat", "OverflowInt", "OverflowUint", "CanSeq", "CanSeq2",
		} {
			if meth.Name == x {
				continue checkMethod
			}
		}
		_, ok := rtyp.MethodByName(meth.Name)
		assert.Truef(t, ok, "missing method %v", meth.Name)
	}
}

func TestRType(t *testing.T) {
	var (
		oneI8    int8  = 1
		twoI32   int32 = 2
		threeInt int   = 3
		strA           = "a"
	)
	types := []any{
		int8(1),
		&oneI8,
		int32(2),
		&twoI32,
		int(3),
		&threeInt,
		"a",
		&strA,
		simple{"b"},
		&simple{"b"},
	}
	for _, x := range types {
		rtype1 := EfaceOf(&x).RType
		rtype2 := RTypeOf(reflect.TypeOf(x))
		rtype3 := RTypeOf(reflect.ValueOf(x))
		assert.Equal(t, rtype1, rtype2)
		assert.Equal(t, rtype2, rtype3)
	}

	assert.Equal(t, reflect.TypeOf(0).Size(), RTypeOf(0).Size())
	assert.Equal(t, reflect.Int, RTypeOf(0).Kind())
	assert.Equal(t, reflect.TypeOf("").Size(), RTypeOf("").Size())
	assert.Equal(t, reflect.String, RTypeOf("").Kind())
}

func TestPtrTo(t *testing.T) {
	var x int64
	typ1 := PtrTo(RTypeOf(x)).ToReflectType()
	typ2 := reflect.PtrTo(reflect.TypeOf(x))
	assert.Equal(t, typ1, typ2)
}

func TestRTypeToType(t *testing.T) {
	typ := RTypeOf(123)
	t1 := reflect.TypeOf(123)
	t2 := typ.ToReflectType()

	if1 := unsafeheader.ToIface(t1)
	if2 := unsafeheader.ToIface(t2)
	assert.True(t, if1.Tab == if2.Tab)
	assert.True(t, if1.Data == if2.Data)
}

func TestToRType(t *testing.T) {
	var x int64
	rtyp1 := RTypeOf(x)
	rtyp2 := ToRType(reflect.TypeOf(x))
	assert.Equal(t, rtyp1, rtyp2)
}

func TestRTypeOf(t *testing.T) {
	x := any(123)
	rtyp := EfaceOf(&x).RType
	assert.Equal(t, rtyp, RTypeOf(123))
	assert.Equal(t, rtyp, RTypeOf(reflect.TypeOf(123)))
	assert.Equal(t, rtyp, RTypeOf(reflect.ValueOf(123)))
	assert.Equal(t, rtyp, RTypeOf(RTypeOf(123)))
}

func BenchmarkRTypeSizeAndKind(b *testing.B) {
	typ := RTypeOf("hello")
	for i := 0; i < b.N; i++ {
		_ = typ.Size()
		_ = typ.Kind()
	}
}

func BenchmarkRTypeSizeAndKind_linkname(b *testing.B) {
	typ := RTypeOf("hello")
	for i := 0; i < b.N; i++ {
		_ = linkname.Reflect_rtype_Size(unsafe.Pointer(typ))
		_ = linkname.Reflect_rtype_Kind(unsafe.Pointer(typ))
	}
}
