package bbp

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLinkBufferWrite(t *testing.T) {
	str := "abc"
	want := ""
	buf := NewLinkBuffer(64)
	for i := 0; i < 1000; i++ {
		tmp := strings.Repeat(str, i)
		want += tmp
		tmpNN, tmpErr := buf.WriteString(tmp)

		assert.Nil(t, tmpErr)
		assert.Equal(t, len(tmp), tmpNN)
	}
	got := buf.String()
	assert.Equal(t, want, got)
}

func TestLinkBufferReadFrom(t *testing.T) {
	str := "abc"
	want := ""
	bytesBuf := bytes.NewBuffer(nil)
	for i := 0; i < 1000; i++ {
		tmp := strings.Repeat(str, i)
		want += tmp
		tmpNN, tmpErr := bytesBuf.WriteString(tmp)

		assert.Nil(t, tmpErr)
		assert.Equal(t, len(tmp), tmpNN)
	}
	bytesStr := bytesBuf.String()

	buf := NewLinkBuffer(64)
	n, err := buf.ReadFrom(bytesBuf)
	assert.Nil(t, err)
	assert.Equal(t, len(want), int(n))
	assert.Equal(t, want, bytesStr)
	assert.Equal(t, want, buf.String())
}
