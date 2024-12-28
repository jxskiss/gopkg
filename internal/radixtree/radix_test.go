package radixtree

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Level = int8

const (
	DebugLevel = iota - 1
	InfoLevel
	WarnLevel
	ErrorLevel
)

func TestRadixTree(t *testing.T) {
	tree := New[Level]()
	tree.Insert("some.module_2.pkg_1", InfoLevel)
	tree.Insert("some.module_2.pkg_2", DebugLevel)
	tree.Insert("some2.filtertest", WarnLevel)
	tree.Insert("some.module_1", ErrorLevel)
	tree.Insert("some.module_1.pkg_1", WarnLevel)

	assert.Equal(t, `some.module_1=2
some.module_1.pkg_1=1
some.module_2.pkg_1=0
some.module_2.pkg_2=-1
some2.filtertest=1`, strings.TrimSpace(tree.Dump("")))

	testcases := []struct {
		Name  string
		Level Level
		Found bool
	}{
		{"some.module_1", ErrorLevel, true},
		{"some.module_1.pkg_0", ErrorLevel, true},
		{"some.module_1.pkg_1", WarnLevel, true},
		{"some.module_2", InfoLevel, false},
		{"some.module_2.pkg_0", InfoLevel, false},
		{"some.module_2.pkg_1", InfoLevel, true},
		{"some.module_2.pkg_2", DebugLevel, true},
		{"some2.filtertest", WarnLevel, true},
		{"some2.filtertest.aaa", WarnLevel, true},
	}
	for _, tc := range testcases {
		level, found := tree.Search(tc.Name)
		assert.Equalf(t, tc.Level, level, "name= %v", tc.Name)
		assert.Equal(t, tc.Found, found, "name= %v", tc.Name)
	}
}
