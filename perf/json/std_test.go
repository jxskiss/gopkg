package json

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mutatingUnmarshaler struct {
	value string
}

func (p *mutatingUnmarshaler) UnmarshalJSON(data []byte) error {
	p.value = string(data)
	data[0] = '9'
	return nil
}

func TestStdUnmarshalFromStringPreservesInput(t *testing.T) {
	s := strings.Repeat("1", 8)
	var dst mutatingUnmarshaler
	err := StdImpl.UnmarshalFromString(s, &dst)
	assert.NoError(t, err)
	assert.Equal(t, "11111111", dst.value)
	assert.Equal(t, "11111111", s)
}
