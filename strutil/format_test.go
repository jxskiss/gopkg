package strutil

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPyFormat_AutoNumber(t *testing.T) {
	str := "some test {} and {test1} {} with {test2}"
	got := PyFormat(str, map[string]interface{}{"test1": "abc", "test2": 123}, "position1", []string{"sliceElem1", "sliceElem2"})
	want := "some test position1 and abc [sliceElem1 sliceElem2] with 123"
	assert.Equal(t, want, got)
}

func TestPyFormat_ManualNumber(t *testing.T) {
	str := "some test {0} and {test1} {1} with {test2} {0}"
	got := PyFormat(str, map[string]interface{}{"test1": "abc", "test2": 123}, "position1", []string{"sliceElem1", "sliceElem2"})
	want := "some test position1 and abc [sliceElem1 sliceElem2] with 123 position1"
	assert.Equal(t, want, got)
}

func TestPyFormat_Struct(t *testing.T) {
	str := "some test {Field1:%s} and {Field2:08d}"
	got := PyFormat(str, &testObject{Field1: "abc", Field2: 123})
	want := "some test abc and 00000123"
	assert.Equal(t, want, got)
}

func TestPyFormat_EscapeWing(t *testing.T) {
	got := PyFormat("{{ some text }}", nil)
	want := "{ some text }"
	assert.Equal(t, want, got)
}

func TestPyFormat_Malformed(t *testing.T) {
	str1 := "some test {} {0} {} {1}"
	got1 := PyFormat(str1, nil, "abc", "123")
	want1 := "some test abc {0} 123 {1}"
	assert.Equal(t, want1, got1)

	str2 := "some test {1} {} {0} {}"
	got2 := PyFormat(str2, nil, "abc", "123")
	want2 := "some test 123 {} abc {}"
	assert.Equal(t, want2, got2)

	str3 := "some test {abc} {{abc}} {key2}"
	got3 := PyFormat(str3, map[string]interface{}{"abc": "123"})
	want3 := "some test 123 {abc} {key2}"
	assert.Equal(t, want3, got3)

	str4 := "some test {Field1} {Field3:%.8f} {private:08d}"
	got4 := PyFormat(str4, &testObject{Field1: "abc", Field2: 123, private: 456})
	want4 := "some test abc {Field3:%.8f} 00000456"
	assert.Equal(t, want4, got4)
}

func TestPyFormat_StringInterfaceMap(t *testing.T) {
	str := "some test {abc} {{abc}} {key2}"
	kwArgs := map[String]interface{}{"abc": "123"}
	got := PyFormat(str, kwArgs)
	want := "some test 123 {abc} {key2}"
	assert.Equal(t, want, got)
}

type String string

type testObject struct {
	Field1  string
	Field2  int64
	private int32
}
