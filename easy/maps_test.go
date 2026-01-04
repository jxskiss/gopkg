package easy

import (
	"strconv"
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

func TestCopyMap(t *testing.T) {
	m := map[int64]bool{1: true, 2: false}

	got1 := CopyMap(m)
	assert.Equal(t, m, got1)

	got2 := CopyMap(m, 10)
	assert.Equal(t, m, got2)
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

	m4 := make(map[string]int)
	for i := 0; i < 120; i++ {
		m4[strconv.Itoa(i)] = i
	}
	got4 := SplitMap(m4, 100)
	assert.Len(t, got4, 2)
	assert.Len(t, got4[0], 100)
	assert.Len(t, got4[1], 20)
}

func TestSplitMapStable(t *testing.T) {
	origMap := make(map[string]int)
	for i := 0; i < 300; i++ {
		origMap[strconv.Itoa(i)] = i
	}

	var gotBatchMaps [][]map[string]int
	batchSize := 20
	for i := 0; i < 5; i++ {
		got := SplitMapStable(origMap, batchSize)
		gotBatchMaps = append(gotBatchMaps, got)
		assert.Len(t, got, 15)
		for j := 0; j < len(got); j++ {
			assert.Len(t, got[j], batchSize)
		}
	}

	got0 := gotBatchMaps[0]
	for i := 1; i < len(gotBatchMaps); i++ {
		for j := 0; j < len(got0); j++ {
			assert.Equalf(t, got0[j], gotBatchMaps[i][j], "got map not equal, i= %d, j= %d", i, j)
		}
	}
}
