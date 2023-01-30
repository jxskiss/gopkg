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

func TestSplitMap(t *testing.T) {
	got1 := SplitMap(map[int64]bool{}, 10)
	assert.Nil(t, got1)

	m2 := map[int64]bool{1: true, 2: true, 3: false}
	got2 := SplitMap(m2, 3)
	assert.Len(t, got2, 1)
	assert.Equal(t, m2, got2[0])

	m3 := map[string]bool{"a": true, "b": true, "c": false, "d": true, "e": true, "f": true, "g": true}
	got3 := SplitMap(m3, 4)
	assert.Len(t, got3, 2)
	assert.Len(t, got3[0], 4)
	assert.Len(t, got3[1], 3)
	for _, k := range []string{"a", "b", "d", "e", "f", "g"} {
		assert.True(t, got3[0][k] || got3[1][k])
	}
	cVal1, ok1 := got3[0]["c"]
	cVal2, ok2 := got3[1]["c"]
	assert.True(t, ok1 || ok2)
	assert.False(t, cVal1 || cVal2)
}
