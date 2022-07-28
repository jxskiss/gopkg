package ptr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntToStringp(t *testing.T) {
	got := IntToStringp(1234)
	assert.Equal(t, "1234", *got)
}
