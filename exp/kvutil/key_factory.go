package kvutil

import (
	"fmt"
	"regexp"
	"strings"
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
	return kf.newSprintfKey(format, argNames...)
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
