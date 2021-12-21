package mcli

import (
	"reflect"
	"testing"
)

func Test_parseFlag_SliceDefaultValue(t *testing.T) {
	var args struct {
		Value []string
	}
	cliTag := "-f, -some-flag"
	defaultValue := "a, b\\,cde\\, 345, fgh"
	f := parseFlag(cliTag, defaultValue, reflect.ValueOf(&args).Elem().Field(0))
	_ = f
	if !reflect.DeepEqual(args.Value, []string{"a", "b,cde, 345", "fgh"}) {
		t.Errorf("set slice value failed, got= %q", args.Value)
	}
}
