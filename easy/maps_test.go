package easy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeMaps(t *testing.T) {
	m1 := map[int64]int64{1: 2, 3: 4, 5: 6}
	m2 := map[int64]int64{7: 8, 9: 10}
	got := MergeMaps(m1, m2)
	assert.Equal(t, 3, len(m1))
	assert.Equal(t, 2, len(m2))
	assert.Equal(t, 5, len(got))
	assert.Equal(t, int64(4), got[3])
	assert.Equal(t, int64(8), got[7])
}

func TestMergeMapsTo(t *testing.T) {
	m1 := map[int64]int64{1: 2, 3: 4, 5: 6}
	m2 := map[int64]int64{7: 8, 9: 10}
	_ = MergeMapsTo(m1, m2)
	assert.Equal(t, 5, len(m1))
	assert.Equal(t, 2, len(m2))
	assert.Equal(t, int64(4), m1[3])
	assert.Equal(t, int64(8), m1[7])
	assert.Equal(t, int64(10), m1[9])

	// merge to a nil map
	var m3 map[int64]int64
	m4 := MergeMapsTo(m3, m1)
	assert.Nil(t, m3)
	assert.Equal(t, 5, len(m4))
	assert.Equal(t, int64(4), m4[3])
	assert.Equal(t, int64(10), m4[9])
}