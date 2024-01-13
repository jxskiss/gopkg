package ezdbg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_logFilter(t *testing.T) {
	type matchPair struct {
		fileName string
		match    bool
	}
	testCases := []struct {
		name  string
		rule  string
		match []matchPair
	}{
		{name: "empty", rule: "",
			match: []matchPair{
				{"gopkg/confr/loader.go", true},
				{"gopkg/easy/json.go", true},
				{"gopkg/easy/ezdbg/debug_test.go", true},
			}},
		{name: "allow all", rule: "allow=all",
			match: []matchPair{
				{"gopkg/confr/loader.go", true},
				{"gopkg/easy/json.go", true},
				{"gopkg/easy/ezdbg/debug_test.go", true},
				{"gopkg/pkg1/a.go", true},
				{"gopkg/pkg1/sub/b.go", true},
				{"gopkg/pkg1/sub/sub/a.go", true},
				{"gopkg/pkg2/a.go", true},
				{"gopkg/pkg2/sub/b.go", true},
			}},
		{name: "allow 1", rule: "allow=pkg1/*,pkg2/*.go",
			match: []matchPair{
				{"gopkg/confr/loader.go", false},
				{"gopkg/easy/json.go", false},
				{"gopkg/easy/ezdbg/debug_test.go", false},
				{"gopkg/pkg1/a.go", true},
				{"gopkg/pkg1/sub/b.go", false},
				{"gopkg/pkg2/a.go", true},
				{"gopkg/pkg2/sub/b.go", false},
			}},
		{name: "allow 2", rule: "allow=pkg1/sub1/abc.go,pkg1/sub2/def.go",
			match: []matchPair{
				{"gopkg/confr/loader.go", false},
				{"gopkg/easy/json.go", false},
				{"gopkg/easy/ezdbg/debug_test.go", false},
				{"gopkg/pkg1/a.go", false},
				{"gopkg/pkg1/sub1/abc.go", true},
				{"gopkg/pkg1/sub2/abc.go", false},
				{"gopkg/pkg1/sub2/def.go", true},
			}},
		{name: "allow 3", rule: "allow=pkg1/**",
			match: []matchPair{
				{"gopkg/confr/loader.go", false},
				{"gopkg/easy/json.go", false},
				{"gopkg/easy/ezdbg/debug_test.go", false},
				{"gopkg/pkg1/a.go", true},
				{"gopkg/pkg1/sub/a.go", true},
				{"gopkg/pkg1/sub/sub/a.go", true},
			}},
		{name: "deny all", rule: "deny=all",
			match: []matchPair{
				{"gopkg/confr/loader.go", false},
				{"gopkg/easy/json.go", false},
				{"gopkg/easy/ezdbg/debug_test.go", false},
				{"gopkg/pkg1/a.go", false},
				{"gopkg/pkg1/sub/b.go", false},
				{"gopkg/pkg1/sub/sub/a.go", false},
				{"gopkg/pkg2/a.go", false},
				{"gopkg/pkg2/sub/b.go", false},
			}},
		{name: "deny 1", rule: "deny=pkg1/**.go,pkg2/**.go",
			match: []matchPair{
				{"gopkg/confr/loader.go", true},
				{"gopkg/easy/json.go", true},
				{"gopkg/easy/ezdbg/debug_test.go", true},
				{"gopkg/pkg1/a.go", false},
				{"gopkg/pkg1/sub/a.go", false},
				{"gopkg/pkg1/sub/sub/a.go", false},
				{"gopkg/pkg2/a.go", false},
				{"gopkg/pkg2/sub/a.go", false},
				{"gopkg/pkg2/sub/sub/a.go", false},
				{"gopkg/pkg3/a.go", true},
				{"gopkg/pkg3/sub/a.go", true},
				{"gopkg/pkg3/sub/sub/a.go", true},
			}},
		{name: "deny 2", rule: "allow=all;deny=pkg/**",
			match: []matchPair{
				{"gopkg/confr/loader.go", false},
				{"gopkg/easy/json.go", false},
				{"gopkg/easy/ezdbg/debug_test.go", false},
				{"gopkg/pkg1/a.go", false},
				{"gopkg/pkg1/sub/a.go", false},
				{"gopkg/pkg1/sub/sub/a.go", false},
				{"gopkg/pkg2/a.go", false},
				{"gopkg/pkg2/sub/a.go", false},
				{"gopkg/pkg2/sub/sub/a.go", false},
				{"gopkg/pkg3/a.go", false},
				{"gopkg/pkg3/sub/a.go", false},
				{"gopkg/pkg3/sub/sub/a.go", false},
				{"mcli/cli.go", true},
				{"mcli/pkg/a.go", false},
				{"mcli/pkg/sub/a.go", false},
				{"mcli/pkg/sub/sub/a.go", false},
			}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filter := newLogFilter(tc.rule)
			for _, pair := range tc.match {
				fileName := "/Users/bytedance/go/src/github.com/jxskiss/" + pair.fileName
				got := filter.Allow(pair.fileName)
				assert.Equalf(t, pair.match, got, "rule: %v, fileName: %v", tc.rule, fileName)
			}
		})
	}
}
