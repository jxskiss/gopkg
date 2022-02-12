package easy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterInt64s(t *testing.T) {
	slice := []int64{1, 2, 3, 4, 5, 6}
	got := FilterInt64s(slice, func(i int) bool {
		return slice[i]%2 == 0
	})
	assert.Equal(t, []int64{2, 4, 6}, got)
}

func TestFilterStrings(t *testing.T) {
	slice := []string{"a", "ab", "abc", "abcd"}
	got := FilterStrings(slice, func(i int) bool {
		return len(slice[i]) > 2
	})
	assert.Equal(t, []string{"abc", "abcd"}, got)
}

func TestDiffInt64s(t *testing.T) {
	a := []int64{1, 2, 3, 4, 5}
	b := []int64{4, 5, 6, 7, 8}
	want1 := []int64{1, 2, 3}
	assert.Equal(t, want1, DiffInt64s(a, b))

	want2 := []int64{6, 7, 8}
	assert.Equal(t, want2, DiffInt64s(b, a))
}

func TestDiffStrings(t *testing.T) {
	a := []string{"1", "2", "3", "4", "5"}
	b := []string{"4", "5", "6", "7", "8"}
	want1 := []string{"1", "2", "3"}
	assert.Equal(t, want1, DiffStrings(a, b))

	want2 := []string{"6", "7", "8"}
	assert.Equal(t, want2, DiffStrings(b, a))
}
