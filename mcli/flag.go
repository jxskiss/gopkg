package mcli

import (
	"flag"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

var (
	flagGetterTyp = reflect.TypeOf((*flag.Getter)(nil)).Elem()
	flagValueTyp  = reflect.TypeOf((*flag.Value)(nil)).Elem()
)

// _flag implements flag.Value.
type _flag struct {
	name        string
	short       string
	description string
	defValue    string
	_value

	hasDefault bool
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
	hasDefault := false
	if defaultValue != "" {
		if rv.Kind() == reflect.Slice {
			hasDefault = true
		} else {
			f.Set(defaultValue)
			hasDefault = !f.isZero()
		}
	}
	f.hasDefault = hasDefault
	return f
}

func (f *_flag) Get() interface{} {
	if f.rv.Type().Implements(flagGetterTyp) {
		return f.rv.Interface().(flag.Getter).Get()
	}
	if f.rv.Addr().Type().Implements(flagGetterTyp) {
		return f.rv.Addr().Interface().(flag.Getter).Get()
	}
	return f.rv.Interface()
}

func (f *_flag) String() string {
	return formatValue(f.rv)
}

func formatValue(rv reflect.Value) string {
	if rv.Type().Implements(flagValueTyp) {
		return rv.Interface().(flag.Value).String()
	}
	if rv.Addr().Type().Implements(flagValueTyp) {
		return rv.Addr().Interface().(flag.Value).String()
	}
	switch rv.Kind() {
	case reflect.Bool:
		return strconv.FormatBool(rv.Bool())
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
		if rv.Type() == reflect.TypeOf(time.Duration(0)) {
			return rv.Interface().(time.Duration).String()
		}
		return strconv.FormatInt(rv.Int(), 10)
	case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(rv.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(rv.Float(), 'g', -1, 64)
	case reflect.String:
		return rv.String()
	case reflect.Slice:
		return formatSliceValues(rv)
	default:
		return ""
	}
}

func formatSliceValues(rv reflect.Value) string {
	out := ""
	n := rv.Len()
	for i := 0; i < n; i++ {
		s := formatValue(rv.Index(i))
		if out != "" {
			out += ", "
		}
		out += strings.ReplaceAll(s, ",", "\\,")
	}
	return out
}

func (f *_flag) Set(s string) error {
	return applyValue(f.rv, s)
}

func applyValue(rv reflect.Value, s string) error {
	if s == "" {
		return nil
	}
	if rv.Type().Implements(flagValueTyp) {
		return rv.Interface().(flag.Value).Set(s)
	}
	if rv.Addr().Type().Implements(flagValueTyp) {
		return rv.Addr().Interface().(flag.Value).Set(s)
	}
	switch rv.Kind() {
	case reflect.Bool:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return err
		}
		rv.SetBool(b)
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
		var i int64
		var d time.Duration
		var err error
		if rv.Type() == reflect.TypeOf(time.Duration(0)) {
			d, err = time.ParseDuration(s)
			i = int64(d)
		} else {
			i, err = strconv.ParseInt(s, 10, 64)
		}
		if err != nil {
			return err
		}
		rv.SetInt(i)
	case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return err
		}
		rv.SetUint(u)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return err
		}
		rv.SetFloat(f)
	case reflect.String:
		rv.SetString(s)
	case reflect.Slice:
		e := reflect.New(rv.Type().Elem()).Elem()
		if err := applyValue(e, s); err != nil {
			return err
		}
		rv.Set(reflect.Append(rv, e))
	default:
		panic(fmt.Sprintf("unspported flag value type: %v", rv.Type()))
	}
	return nil
}

func (f *_flag) isSlice() bool {
	return f.rv.Kind() == reflect.Slice
}

func (f *_flag) isString() bool {
	return f.rv.Kind() == reflect.String
}

func (f *_flag) isZero() bool {
	typ := f.rv.Type()
	if f.rv.Type().Comparable() {
		return reflect.Zero(typ).Interface() == f.rv.Interface()
	}
	// else it must be a slice
	return f.rv.Len() == 0
}

func (f *_flag) usageName() string {
	if f.rv.Kind() == reflect.Bool {
		return ""
	}
	if isFlagValueImpl(f.rv) {
		return "value"
	}
	return usageName(f.rv.Type())
}

func usageName(typ reflect.Type) string {
	switch typ.Kind() {
	case reflect.Bool:
		return "bool"
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
		if typ == reflect.TypeOf(time.Duration(0)) {
			return "duration"
		}
		return "int"
	case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "uint"
	case reflect.Float32, reflect.Float64:
		return "float"
	case reflect.String:
		return "string"
	case reflect.Slice:
		elemName := usageName(typ.Elem())
		return "[]" + elemName
	default:
		return "value"
	}
}

func (f *_flag) getUsage(hasShortFlag bool) (prefix, usage string) {
	if f.nonflag {
		prefix += "  " + f.name
	} else if f.short != "" && f.name != "" {
		prefix += fmt.Sprintf("  -%s, --%s", f.short, f.name)
	} else if len(f.name) == 1 || !hasShortFlag {
		prefix += fmt.Sprintf("  -%s", f.name)
	} else {
		prefix += fmt.Sprintf("      --%s", f.name)
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
	if f.hasDefault {
		if f.isString() || f.isSlice() {
			usage += fmt.Sprintf(" (default %q)", f.defValue)
		} else {
			usage += fmt.Sprintf(" (default %v)", f.defValue)
		}
	}
	return
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

func parseTags(fs *flag.FlagSet, rv reflect.Value) (flags, nonflags []*_flag, unsupportedType reflect.Type) {
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		fv := rv.Field(i)
		ft := rt.Field(i)
		cliTag := strings.TrimSpace(ft.Tag.Get("cli"))
		defaultValue := strings.TrimSpace(ft.Tag.Get("default"))
		if isIgnoreTag(cliTag) {
			continue
		}
		if ft.PkgPath != "" { // unexported fields
			continue
		}
		if fv.Kind() == reflect.Struct && !isFlagValueImpl(fv) {
			subFlags, subNonflags, subErr := parseTags(fs, fv)
			if subErr != nil {
				return nil, nil, subErr
			}
			flags = append(flags, subFlags...)
			nonflags = append(nonflags, subNonflags...)
			continue
		}
		if cliTag == "" {
			continue
		}
		if !isSupportedType(fv) {
			return nil, nil, fv.Type()
		}
		f := parseFlag(cliTag, defaultValue, fv)
		if f == nil || f.name == "" {
			continue
		}
		if f.nonflag {
			nonflags = append(nonflags, f)
			continue
		}
		flags = append(flags, f)
		if fv.Kind() == reflect.Bool {
			ptr := fv.Addr().Interface().(*bool)
			fs.BoolVar(ptr, f.name, f.rv.Bool(), f.description)
			if f.short != "" {
				fs.BoolVar(ptr, f.short, f.rv.Bool(), f.description)
			}
			continue
		}
		fs.Var(f, f.name, f.description)
		if f.short != "" {
			fs.Var(f, f.short, f.description)
		}
	}
	sort.Slice(flags, func(i, j int) bool {
		return strings.ToLower(flags[i].name) < strings.ToLower(flags[j].name)
	})
	return
}

func isIgnoreTag(tag string) bool {
	parts := strings.Split(tag, ",")
	return strings.TrimSpace(parts[0]) == "-"
}

func isSupportedType(rv reflect.Value) bool {
	if _, ok := rv.Interface().(bool); ok {
		return true
	}
	if isFlagValueImpl(rv) {
		return true
	}
	if isSupportedBasicType(rv.Kind()) {
		return true
	}
	if rv.Kind() == reflect.Slice && isSupportedBasicType(rv.Type().Elem().Kind()) {
		return true
	}
	return false
}

func isFlagValueImpl(rv reflect.Value) bool {
	return rv.Type().Implements(flagValueTyp) || rv.Addr().Type().Implements(flagValueTyp)
}

func isSupportedBasicType(kind reflect.Kind) bool {
	switch kind {
	case reflect.Bool,
		reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.String:
		return true
	}
	return false
}

var spaceRE = regexp.MustCompile(`\s+`)

func parseFlag(cliTag, defaultValue string, rv reflect.Value) *_flag {
	f := newFlag(rv, defaultValue)

	const (
		modifier = iota
		short
		long
		description
		stop
	)
	parts := strings.SplitN(cliTag, ",", 4)
	st := modifier
	for i := 0; i < len(parts) && st < stop; i++ {
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
			if strings.HasPrefix(p, "-") {
				st = long
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
			st = stop
			p = strings.TrimSpace(strings.Join(parts[i:], ","))
			f.description = p
		}
	}
	if f.name == "" {
		f.name = f.short
	}
	if f.short == f.name {
		f.short = ""
	}
	return f
}

func splitSliceValues(value string) []string {
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

var (
	_flagSetActualOffset uintptr
	_flagSetFormalOffset uintptr
	_flagSetMapType      = reflect.TypeOf(map[string]*flag.Flag{})
)

func init() {
	typ := reflect.TypeOf(flag.FlagSet{})
	actualField, ok1 := typ.FieldByName("actual")
	formalField, ok2 := typ.FieldByName("formal")
	if !ok1 || !ok2 {
		panic("cannot find flag.FlagSet fields actual/formal")
	}
	if actualField.Type != _flagSetMapType || formalField.Type != _flagSetMapType {
		panic("type of flag.FlagSet fields actual/formal is not map[string]*flag.Flag")
	}
	_flagSetActualOffset = actualField.Offset
	_flagSetFormalOffset = formalField.Offset
}

func _flagSet_getActual(fs *flag.FlagSet) map[string]*flag.Flag {
	return *(*map[string]*flag.Flag)(unsafe.Pointer(uintptr(unsafe.Pointer(fs)) + _flagSetActualOffset))
}

func _flagSet_getFormal(fs *flag.FlagSet) map[string]*flag.Flag {
	return *(*map[string]*flag.Flag)(unsafe.Pointer(uintptr(unsafe.Pointer(fs)) + _flagSetFormalOffset))
}
