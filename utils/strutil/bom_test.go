package strutil

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectBOM(t *testing.T) {
	tests := []struct {
		b    []byte
		want string
	}{
		{[]byte{0xEF, 0xBB, 0xBF, 'a', 'b', 'c'}, BOM_UTF8},
		{[]byte{0xFE, 0xFF, 'a', 'b', 'c'}, BOM_UTF16_BigEndian},
		{[]byte{0xFF, 0xFE, 'a', 'b', 'c'}, BOM_UTF16_LittleEndian},
		{[]byte{0x00, 0x00, 0xFE, 0xFF, 'a', 'b', 'c'}, BOM_UTF32_BigEndian},
		{[]byte{0xFF, 0xFE, 0x00, 0x00, 'a', 'b', 'c'}, BOM_UTF32_LittleEndian},
	}

	for _, x := range tests {
		got := DetectBOM(x.b)
		assert.Equal(t, x.want, got)
	}
}

func TestTrimBOM(t *testing.T) {
	tests := [][]byte{
		{0xEF, 0xBB, 0xBF, 'a', 'b', 'c'},
		{0xFE, 0xFF, 'a', 'b', 'c'},
		{0xFF, 0xFE, 'a', 'b', 'c'},
		{0x00, 0x00, 0xFE, 0xFF, 'a', 'b', 'c'},
		{0xFF, 0xFE, 0x00, 0x00, 'a', 'b', 'c'},
	}
	want := []byte{'a', 'b', 'c'}
	for _, x := range tests {
		got := TrimBOM(x)
		assert.Equal(t, want, got)
	}
}

func TestSkipBOM(t *testing.T) {
	tests := [][]byte{
		{0xEF, 0xBB, 0xBF, 'a', 'b', 'c'},
		{0xFE, 0xFF, 'a', 'b', 'c'},
		{0xFF, 0xFE, 'a', 'b', 'c'},
		{0x00, 0x00, 0xFE, 0xFF, 'a', 'b', 'c'},
		{0xFF, 0xFE, 0x00, 0x00, 'a', 'b', 'c'},
	}
	want := []byte{'a', 'b', 'c'}
	for _, x := range tests {
		rd := bytes.NewBuffer(x)
		tmp := SkipBOMReader(rd)
		got, err := ioutil.ReadAll(tmp)
		assert.Nil(t, err)
		assert.Equal(t, want, got)
	}

	tests = [][]byte{
		{0xEF, 0xBB, 0xBF},
		{0xFE, 0xFF},
		{0xFF, 0xFE},
		{0x00, 0x00, 0xFE, 0xFF},
		{0xFF, 0xFE, 0x00, 0x00},
	}
	want = []byte{}
	for _, x := range tests {
		rd := bytes.NewBuffer(x)
		tmp := SkipBOMReader(rd)
		got, err := ioutil.ReadAll(tmp)
		assert.Nil(t, err)
		assert.Equal(t, want, got)
	}
}
