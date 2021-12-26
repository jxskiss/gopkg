package mcli

import (
	"bytes"
	"flag"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_splitSliceValues(t *testing.T) {
	resetState()
	var args struct {
		Value []string
	}
	defaultValue := "a, b\\,cde\\, 345, fgh"
	rv := reflect.ValueOf(&args).Elem().Field(0)
	for _, s := range splitSliceValues(defaultValue) {
		err := applyValue(rv, s)
		if err != nil {
			t.Errorf("applyValue slice got err, %v", err)
		}
	}
	if !reflect.DeepEqual(args.Value, []string{"a", "b,cde, 345", "fgh"}) {
		t.Errorf("set slice value failed, got= %q", args.Value)
	}
}

func Test_flag_DefaultValue(t *testing.T) {
	resetState()
	var args struct {
		A bool          `cli:"-a" default:"true"`
		B string        `cli:"-b" default:"astr"`
		C []string      `cli:"-c" default:"a, b\\,cde\\, 345, fgh"`
		D time.Duration `cli:"-d" default:"1.5s"`

		Arg1 string   `cli:"arg1" default:"arg1str"`
		Arg2 string   `cli:"arg2" default:"arg2str"`
		Arg3 []string `cli:"arg3" default:"a, b, c"`
	}
	fs, err := Parse(&args, WithErrorHandling(flag.ContinueOnError),
		WithArgs([]string{"-d", "1000ms", "cmdlineArg1"}))
	assert.Nil(t, err)
	assert.Equal(t, true, args.A)
	assert.Equal(t, "astr", args.B)
	assert.Equal(t, []string{"a", "b,cde, 345", "fgh"}, args.C)
	assert.Equal(t, 1000*time.Millisecond, args.D)
	assert.Equal(t, "cmdlineArg1", args.Arg1)
	assert.Equal(t, "arg2str", args.Arg2)
	assert.Equal(t, []string{"a", "b", "c"}, args.Arg3)

	var buf bytes.Buffer
	fs.SetOutput(&buf)
	fs.Usage()

	got := buf.String()
	assert.Contains(t, got, "FLAGS:")
	assert.Contains(t, got, "  -a")
	assert.Contains(t, got, "(default true)")
	assert.Contains(t, got, "  -b string")
	assert.Contains(t, got, `(default "astr")`)
	assert.Contains(t, got, "  -c []string")
	assert.Contains(t, got, `(default "a, b\\,cde\\, 345, fgh")`)
	assert.Contains(t, got, "  -d duration")
	assert.Contains(t, got, "(default 1.5s)")
	assert.Contains(t, got, "ARGUMENTS:")
	assert.Contains(t, got, "  arg1 string")
	assert.Contains(t, got, `(default "arg1str")`)
	assert.Contains(t, got, "  arg2 string")
	assert.Contains(t, got, `(default "arg2str")`)
	assert.Contains(t, got, "  arg3 []string")
	assert.Contains(t, got, `(default "a, b, c")`)
}
