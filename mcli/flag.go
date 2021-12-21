package mcli

import (
	"flag"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// _flag implements flag.Value.
type _flag struct {
	name        string
	short       string
	description string
	defValue    string
	_value

	deprecated bool
	hidden     bool
	required   bool
	nonflag    bool
}

type _value struct {
	rv reflect.Value
}

func newFlag(rv reflect.Value, defaultValue string) *_flag {
	f := &_flag{
		defValue: defaultValue,
		_value:   _value{rv},
	}
	if defaultValue != "" {
		if rv.Kind() == reflect.Slice {
			for _, value := range splitSliceDefaultValues(defaultValue) {
				f.Set(value)
			}
		} else {
			f.Set(defaultValue)
		}
	}
	return f
}

func (f *_flag) String() string {
	// The return value is not used in this library.
	return f.defValue
}

func (f *_flag) Set(s string) error {
	return applyValue(f.rv, s)
}

func (f *_flag) isSlice() bool {
	return f.rv.Kind() == reflect.Slice
}

func (f *_flag) isString() bool {
	return f.rv.Kind() == reflect.String
}

func (f *_flag) isZero() bool {
	typ := f.rv.Type()
	return reflect.Zero(typ).Interface() == f.rv.Interface()
}

func (f *_flag) usageName() string {
	if f.rv.Kind() == reflect.Bool {
		return ""
	}
	return usageName(f.rv.Type())
}

func (f *_flag) getUsage(hasShortFlag bool) (prefix, usage string) {
	if f.nonflag {
		prefix += " " + f.name
	} else if f.short != "" && f.name != "" {
		prefix += fmt.Sprintf("  -%s, -%s", f.short, f.name)
	} else if len(f.name) == 1 || !hasShortFlag {
		prefix += fmt.Sprintf("  -%s", f.name)
	} else {
		prefix += fmt.Sprintf("      -%s", f.name)
	}
	name, usage := unquoteUsage(f)
	if name != "" {
		prefix += " " + name
	}
	if f.required {
		prefix += " (REQUIRED)"
	} else if f.deprecated {
		prefix += " (DEPRECATED)"
	}
	if !f.isZero() {
		if f.isString() {
			usage += fmt.Sprintf(" (default %q)", f.defValue)
		} else {
			usage += fmt.Sprintf(" (default %v)", f.defValue)
		}
	}
	return
}

var flagValueTyp = reflect.TypeOf((*flag.Value)(nil)).Elem()

func applyValue(v reflect.Value, s string) error {
	if s == "" {
		return nil
	}
	if v.Type().Implements(flagValueTyp) {
		return v.Interface().(flag.Value).Set(s)
	}
	if v.Addr().Type().Implements(flagValueTyp) {
		return v.Addr().Interface().(flag.Value).Set(s)
	}
	switch v.Kind() {
	case reflect.Bool:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return err
		}
		v.SetBool(b)
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
		var i int64
		var d time.Duration
		var err error
		if v.Type() == reflect.TypeOf(time.Duration(0)) {
			d, err = time.ParseDuration(s)
			i = int64(d)
		} else {
			i, err = strconv.ParseInt(s, 10, 64)
		}
		if err != nil {
			return err
		}
		v.SetInt(i)
	case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return err
		}
		v.SetUint(u)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return err
		}
		v.SetFloat(f)
	case reflect.String:
		v.SetString(s)
	case reflect.Slice:
		e := reflect.New(v.Type().Elem()).Elem()
		if err := applyValue(e, s); err != nil {
			return err
		}
		v.Set(reflect.Append(v, e))
	}
	return nil
}

func usageName(typ reflect.Type) string {
	switch typ.Kind() {
	case reflect.Bool:
		return "bool"
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
		return "int"
	case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "uint"
	case reflect.Float32, reflect.Float64:
		return "float"
	case reflect.String:
		return "string"
	case reflect.Slice:
		elemName := usageName(typ.Elem())
		return "list of " + elemName
	default:
		return "value"
	}
}

func unquoteUsage(f *_flag) (name, usage string) {
	usage = f.description
	for i := 0; i < len(usage); i++ {
		if usage[i] == '`' || usage[i] == '\'' {
			c := usage[i]
			for j := i + 1; j < len(usage); j++ {
				if usage[j] == c {
					name = usage[i+1 : j]
					usage = usage[:i] + name + usage[j+1:]
					return name, usage
				}
			}
			break // Only one back quote; use type name.
		}
	}
	if name == "" {
		name = f.usageName()
	}
	return
}

var spaceRE = regexp.MustCompile(`\s+`)

func parseFlag(cliTag, defaultValue string, rv reflect.Value) *_flag {
	if isIgnoreTag(cliTag) {
		return nil
	}
	f := newFlag(rv, defaultValue)

	const (
		modifier = iota
		short
		long
		description
	)
	parts := strings.SplitN(cliTag, ",", 4)
	st := modifier
	for i := 0; i < len(parts); i++ {
		p := strings.TrimSpace(parts[i])
		switch st {
		case modifier:
			st = short
			if strings.HasPrefix(p, "#") {
				for _, x := range p[1:] {
					switch x {
					case 'D':
						f.deprecated = true
					case 'H':
						f.hidden = true
					case 'R':
						f.required = true
					}
				}
				continue
			}
			i--
		case short:
			st = long
			if strings.HasPrefix(p, "-") {
				p = strings.TrimLeft(p, "-")
				if len(p) == 1 {
					f.short = p
				} else {
					i--
				}
			} else {
				st = description
				f.nonflag = true
				f.name = p
			}
		case long:
			st = description
			if strings.HasPrefix(p, "-") {
				p = strings.TrimLeft(p, "-")
				// Allow split flag name and description by spaces.
				sParts := spaceRE.Split(p, 2)
				f.name = sParts[0]
				newParts := append(parts[:i:i], sParts...)
				newParts = append(newParts, parts[i+1:]...)
				parts = newParts
				continue
			}
			f.name = f.short
			i--
		case description:
			p = strings.TrimSpace(strings.Join(parts[i:], ","))
			f.description = p
			return f
		}
	}
	if f.short == f.name {
		f.short = ""
	}
	return f
}

func isIgnoreTag(tag string) bool {
	parts := strings.Split(tag, ",")
	return strings.TrimSpace(parts[0]) == "-"
}

func splitSliceDefaultValues(value string) []string {
	var out []string
	var p string
	parts := strings.Split(value, ",")
	for i := 0; i < len(parts); i++ {
		p += parts[i]
		if p == "" {
			continue
		}
		for p[len(p)-1] == '\\' && i < len(parts)-1 {
			p = p[:len(p)-1] + "," + parts[i+1]
			i++
		}
		out = append(out, strings.TrimSpace(p))
		p = ""
	}
	return out
}
