package validat

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseInt64(t *testing.T) {
	got1, err := parseInt64(int64(123))
	assert.Nil(t, err)
	assert.Equal(t, int64(123), got1)

	got2, err := parseInt64("123")
	assert.Nil(t, err)
	assert.Equal(t, int64(123), got2)
}

func TestParseInt64s(t *testing.T) {
	got1, err := parseInt64s([]int64{4, 5, 6})
	assert.Nil(t, err)
	assert.Equal(t, []int64{4, 5, 6}, got1)

	got2, err := parseInt64s([]string{"7", "8", "9"})
	assert.Nil(t, err)
	assert.Equal(t, []int64{7, 8, 9}, got2)
}
