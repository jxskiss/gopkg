package kvutil

import (
	"fmt"
	"regexp"
	"strings"
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
