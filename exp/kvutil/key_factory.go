package kvutil

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/jxskiss/gopkg/v2/unsafe/reflectx"
)

var argPattern = regexp.MustCompile(`\{[^{]*\}`)

// Key is a function which formats arguments to a string key using
// predefined key format.
type Key func(args ...any) string

// KeyFactory builds Key functions.
type KeyFactory struct {
	prefix string
}

// SetPrefix configures the Key functions created by the factory
// to add a prefix to all generated cache keys.
func (kf *KeyFactory) SetPrefix(prefix string) {
	kf.prefix = prefix
}

// NewKey creates a Key function.
//
// If argNames are given (eg. arg1, arg2), it replace the placeholders of
// `{argN}` in format to "%v" as key arguments, else it uses a regular
// expression `\{[^{]*\}` to replace all placeholders of `{arg}` in format
// to "%v" as key arguments.
func (kf *KeyFactory) NewKey(format string, argNames ...string) Key {
	if strings.Contains(format, "%") {
		return kf.newSprintfKey(format, argNames...)
	}
	return kf.newBuilderKey(format, argNames...)
}

func (kf *KeyFactory) newSprintfKey(format string, argNames ...string) Key {
	var tmpl string
	if len(argNames) == 0 {
		tmpl = argPattern.ReplaceAllString(format, "%v")
	} else {
		var oldnew []string
		for _, arg := range argNames {
			placeholder := fmt.Sprintf("{%s}", arg)
			oldnew = append(oldnew, placeholder, "%v")
		}
		tmpl = strings.NewReplacer(oldnew...).Replace(format)
	}
	return func(args ...any) string {
		return kf.prefix + fmt.Sprintf(tmpl, args...)
	}
}

// newBuilderKey gives slightly better performance than newSprintfKey.
func (kf *KeyFactory) newBuilderKey(format string, argNames ...string) Key {
	var tmpl, vars []string
	if len(argNames) == 0 {
		tmpl = argPattern.Split(format, -1)
		vars = argPattern.FindAllString(format, -1)
	} else {
		re := ""
		for i, x := range argNames {
			if i > 0 {
				re += "|"
			}
			re += regexp.QuoteMeta("{" + x + "}")
		}
		exp := regexp.MustCompile(re)
		tmpl = exp.Split(format, -1)
		vars = exp.FindAllString(format, -1)
	}
	return func(args ...any) string {
		return buildKey(kf.prefix, tmpl, vars, args)
	}
}

func buildKey(prefix string, tmpl, vars []string, args []any) string {
	buf := make([]byte, 0, 128)
	buf = append(buf, prefix...)
	var i int
	for i = 0; i < len(vars) && i < len(args); i++ {
		buf = append(buf, tmpl[i]...)
		argef := reflectx.EfaceOf(&args[i])
		kind, data := argef.RType.Kind(), argef.Word
		switch kind {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			val := reflectx.CastIntPointer(kind, data)
			buf = strconv.AppendInt(buf, val, 10)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			val := reflectx.CastIntPointer(kind, data)
			buf = strconv.AppendUint(buf, uint64(val), 10)
		case reflect.String:
			buf = append(buf, *(*string)(data)...)
		default:
			buf = append(buf, fmt.Sprint(args[i])...)
		}
	}
	if i < len(vars) {
		for ; i < len(vars); i++ {
			buf = append(buf, tmpl[i]...)
			buf = append(buf, vars[i]...)
		}
	}
	buf = append(buf, tmpl[i]...)
	return reflectx.BytesToString(buf)
}