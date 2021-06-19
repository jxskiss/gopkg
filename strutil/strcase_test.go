package strutil

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestToCamelCase(t *testing.T) {
	cases := [][]string{
		{"test_case", "TestCase"},
		{"test.case", "TestCase"},
		{"test", "Test"},
		{"TestCase", "TestCase"},
		{" test  case ", "TestCase"},
		{"", ""},
		{"many_many_words", "ManyManyWords"},
		{"AnyKind of_string", "AnyKindOfString"},
		{"odd-fix", "OddFix"},
		{"numbers2And55with000", "Numbers2And55With000"},
		{"id", "Id"},
		{"ID", "ID"},
		{"someID", "SomeID"},
		{"someHTMLWord", "SomeHTMLWord"},
	}
	for _, tc := range cases {
		want := tc[1]
		got := ToCamelCase(tc[0])
		assert.Equal(t, want, got)
	}
}

func TestToLowerCamelCase(t *testing.T) {
	cases := [][]string{
		{"test_case", "testCase"},
		{"ID", "id"},
		{"api-example", "apiExample"},
		{"APIExample", "apiExample"},
		{"ILoveYou", "iLoveYou"},
	}
	for _, tc := range cases {
		want := tc[1]
		got := ToLowerCamelCase(tc[0])
		assert.Equal(t, want, got)
	}
}
