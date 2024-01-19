package yamlx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_hasTostrModifier(t *testing.T) {
	testCases := []struct {
		path     string
		isTostr  bool
		modifier string
	}{
		{"0.step.script.0", false, ""},
		{"step|@tostr", true, "|@tostr"},
		{`friends.#(last="Murphy")#.@tostr`, true, ".@tostr"},
		{`friends.#(last="Murphy")#.@tostr.@reverse`, true, ".@tostr.@reverse"},
	}
	for _, tc := range testCases {
		ok, modifier := hasTostrModifier(tc.path)
		assert.Equalf(t, tc.isTostr, ok, "path= %v", tc.path)
		assert.Equalf(t, tc.modifier, modifier, "path= %v", tc.path)
	}
}
