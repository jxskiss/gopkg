package set

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInt(t *testing.T) {
	s := NewInt(1, 3, 5)
	_ = NewIntWithSize(3) // coverage
	assert.Equal(t, 3, s.Size())
	assert.True(t, s.Contains(1))
	assert.False(t, s.Contains(2))
}

func TestInt64(t *testing.T) {
	s := NewInt64(1, 3, 5)
	_ = NewInt64WithSize(3) // coverage
	assert.Equal(t, 3, s.Size())
	assert.True(t, s.Contains(1))
	assert.False(t, s.Contains(2))
}

func TestInt32(t *testing.T) {
	s := NewInt32(1, 3, 5)
	_ = NewInt32WithSize(3) // coverage
	assert.Equal(t, 3, s.Size())
	assert.True(t, s.Contains(1))
	assert.False(t, s.Contains(2))
}

func TestString(t *testing.T) {
	s1 := NewString("a", "b", "c")
	s2 := NewStringWithSize(4)
	slice := []string{"c", "d", "e"}
	s2.Add(slice...)

	assert.Equal(t, 2, s1.Diff(s2).Size())
	assert.Equal(t, 2, s1.DiffSlice(slice).Size())
	assert.Equal(t, 1, s1.Intersect(s2).Size())
	assert.Equal(t, 1, s1.IntersectSlice(slice).Size())
	assert.Equal(t, 5, s1.Union(s2).Size())
	assert.Equal(t, 5, s1.UnionSlice(slice).Size())

	s1.Del("a")
	assert.ElementsMatch(t, []string{"b", "c"}, s1.Slice())
	assert.ElementsMatch(t, []string{"c"}, s1.FilterInclude(slice))
	assert.ElementsMatch(t, []string{"d", "e"}, s1.FilterExclude(slice))
}
