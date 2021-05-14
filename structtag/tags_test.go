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
				{raw: "type:varchar(100)", k: "type", v: "varchar(100)"},
				{raw: "unique_index", k: "unique_index", v: ""},
			},
		},
		{
			tag: `unique;not null`,
			options: Options{
				{raw: "unique", k: "unique", v: ""},
				{raw: "not null", k: "not null", v: ""},
			},
		},
		{
			tag:     `AUTO_INCREMENT`,
			options: Options{{raw: "AUTO_INCREMENT", k: "AUTO_INCREMENT", v: ""}},
		},
		{
			tag:     `-`,
			options: Options{{raw: "-", k: "-", v: ""}},
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
	options := Options{{k: "type", v: "varchar(100)"}, {k: "unique_index", v: ""}}
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
			assert.Equal(t, ts.exp, opt.Value())
		}
	}
}
