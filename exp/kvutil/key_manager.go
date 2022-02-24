package kvutil

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/jxskiss/gopkg/v2/reflectx"
)

var argPattern = regexp.MustCompile(`\{[^{]*\}`)

// Key is a function which formats arguments to a string key using
// predefined key format.
type Key func(args ...interface{}) string

// KeyManager provides utilities to work with cache keys.
type KeyManager struct {
	prefix string
}

// SetPrefix configures the manager using the given prefix to generate
// cache keys.
func (km *KeyManager) SetPrefix(prefix string) {
	km.prefix = prefix
}

// NewKey returns a function to generate cache keys.
//
// If argNames are given (eg. arg1, arg2), it replace the placeholders of
// `{argN}` in format to "%v" as key arguments, else it uses a regular
// expression `\{[^{]*\}` to replace all placeholders of `{arg}` in format
// to "%v" as key arguments.
func (km *KeyManager) NewKey(format string, argNames ...string) Key {
	if strings.Contains(format, "%") {
		return km.newSprintfKey(format, argNames...)
	}
	return km.newBuilderKey(format, argNames...)
}

func (km *KeyManager) newSprintfKey(format string, argNames ...string) Key {
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
	return func(args ...interface{}) string {
		return km.prefix + fmt.Sprintf(tmpl, args...)
	}
}

// newBuilderKey gives better performance than newSprintfKey.
func (km *KeyManager) newBuilderKey(format string, argNames ...string) Key {
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
	return func(args ...interface{}) string {
		return buildKey(km.prefix, tmpl, vars, args)
	}
}

func buildKey(prefix string, tmpl, vars []string, args []interface{}) string {
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
