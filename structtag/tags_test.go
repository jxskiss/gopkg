package structtag

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseOptions_Gorm(t *testing.T) {
	gormTests := []struct {
		tag     string
		options Options
	}{
		{
			tag: `type:varchar(100);unique_index`,
			options: Options{
				{Value: "type:varchar(100)", K: "type", V: "varchar(100)"},
				{Value: "unique_index", K: "unique_index"},
			},
		},
		{
			tag: `unique;not null`,
			options: Options{
				{Value: "unique", K: "unique"},
				{Value: "not null", K: "not null"},
			},
		},
		{
			tag:     `AUTO_INCREMENT`,
			options: Options{{Value: "AUTO_INCREMENT", K: "AUTO_INCREMENT"}},
		},
		{
			tag:     `-`,
			options: Options{{Value: "-", K: "-"}},
		},
		{
			tag:     "",
			options: nil,
		},
	}

	for _, ts := range gormTests {
		options := ParseOptions(ts.tag, ";", ":")
		assert.Equal(t, ts.options, options)
	}
}

func TestOptions_Get(t *testing.T) {
	options := Options{{K: "type", V: "varchar(100)"}, {K: "unique_index"}}
	tests := []struct {
		opt   string
		exp   string
		found bool
	}{
		{
			opt:   "type",
			exp:   "varchar(100)",
			found: true,
		},
		{
			opt:   "unique_index",
			exp:   "",
			found: true,
		},
		{
			opt:   "column",
			found: false,
		},
	}

	for _, ts := range tests {
		opt, found := options.Get(ts.opt)
		assert.Equal(t, ts.found, found)
		if found {
			assert.Equal(t, ts.exp, opt)
		}
	}
}
