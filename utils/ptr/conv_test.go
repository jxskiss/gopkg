package ptr

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntToStringp(t *testing.T) {
	got1 := IntToStringp(1234)
	assert.Equal(t, "1234", *got1)

	want2 := fmt.Sprint(uint64(math.MaxUint64))
	got2 := IntToStringp(uint64(math.MaxUint64))
	assert.Equal(t, want2, *got2)
}

func TestIntpToStringp(t *testing.T) {
	got1 := IntpToStringp(Int32(1234))
	assert.Equal(t, "1234", *got1)

	want2 := fmt.Sprint(uint64(math.MaxUint64))
	got2 := IntpToStringp(Uint64(uint64(math.MaxUint64)))
	assert.Equal(t, want2, *got2)
}

func TestIntpToString(t *testing.T) {
	got1 := IntpToString(Int32(1234))
	assert.Equal(t, "1234", got1)

	want2 := fmt.Sprint(uint64(math.MaxUint64))
	got2 := IntpToString(Uint64(uint64(math.MaxUint64)))
	assert.Equal(t, want2, got2)
}
