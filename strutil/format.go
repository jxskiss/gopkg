package strutil

import (
	"fmt"
	"reflect"
	"strconv"
	"unsafe"

	"github.com/jxskiss/gopkg/reflectx"
)

// Format mimics subset features of the python string.format function.
//
// It formats the string using given keyword arguments and positional arguments.
// kwArgs can be a map[string]interface{}, map[string]string or a struct
// or a pointer to a struct.
//
// If bracket is needed in string it can be created by escaping (two brackets).
//
// All standard formatting options from fmt work. To specify them, add colon
// after key name or position number and specify fmt package compatible formatting
// options. The percent sign is optional. For example:
//
//   // returns "3.14, 3.1416"
//   Format("{pi:%.2f}, {pi:.4f}", map[string]interface{}{"pi": math.Pi})
//
// If a replacement is not found in kwArgs and posArgs, the placeholder will be
// output as the same in the given format.
func Format(format string, kwArgs interface{}, posArgs ...interface{}) string {
	var (
		defaultFormat = []rune("%v")

		// newFormat holds the new format string
		newFormat     = make([]rune, 0, len(format))
		newFormatArgs []interface{}

		prevChar      rune
		currentName   = make([]rune, 0, 10)
		currentFormat = make([]rune, 0, 10)

		inWing      bool
		inWingParam bool

		isAutoNumber   bool
		isManualNumber bool
		argIndex       int

		kwGetter = getKeywordArgFunc(kwArgs)
	)

	for i, char := range format {
		if i > 0 {
			prevChar = rune(format[i-1])
		}
		switch char {
		case '{':
			if inWing && prevChar == '{' {
				inWing = false
				newFormat = append(newFormat, char)
				break
			}
			inWing = true
		case '}':
			if !inWing {
				if prevChar == '}' {
					newFormat = append(newFormat, char)
				}
				break
			}
			isInvalid := false

			// find the argument
			name := string(currentName)
			if name == "" {
				if isManualNumber || argIndex > len(posArgs) {
					isInvalid = true
				} else {
					arg := posArgs[argIndex]
					newFormatArgs = append(newFormatArgs, arg)
					argIndex++
					isAutoNumber = true
				}
			} else if IsASCIIDigit(name) {
				argNum, _ := strconv.ParseInt(name, 10, 64)
				if isAutoNumber || int(argNum) >= len(posArgs) {
					isInvalid = true
				} else {
					arg := posArgs[argNum]
					newFormatArgs = append(newFormatArgs, arg)
					isManualNumber = true
				}
			} else {
				arg, ok := kwGetter(name)
				if !ok {
					isInvalid = true
				} else {
					newFormatArgs = append(newFormatArgs, arg)
				}
			}
			if isInvalid {
				newFormat = append(newFormat, '{')
				newFormat = append(newFormat, currentName...)
				if len(currentFormat) > 0 {
					newFormat = append(newFormat, ':')
					newFormat = append(newFormat, '%')
					newFormat = append(newFormat, currentFormat...)
				}
				newFormat = append(newFormat, '}')
			} else {
				if len(currentFormat) > 0 {
					newFormat = append(newFormat, currentFormat...)
				} else {
					newFormat = append(newFormat, defaultFormat...)
				}
			}
			currentName = currentName[:0]
			currentFormat = currentFormat[:0]

			inWing = false
			inWingParam = false
		case ':':
			if inWing {
				inWingParam = true
			}
		default:
			if inWing {
				if inWingParam {
					if prevChar == ':' && char != '%' {
						currentFormat = append(currentFormat, '%')
					}
					currentFormat = append(currentFormat, char)
				} else {
					currentName = append(currentName, char)
				}
			} else {
				newFormat = append(newFormat, char)
			}
		}
	}

	return fmt.Sprintf(string(newFormat), newFormatArgs...)
}

var strInterfaceMapTyp = reflect.TypeOf(map[string]interface{}(nil))

func isStringInterfaceMap(typ reflect.Type) bool {
	return typ.Kind() == reflect.Map &&
		typ.Key().Kind() == reflect.String &&
		typ.Elem() == strInterfaceMapTyp.Elem()
}

func castStringInterfaceMap(v interface{}) map[string]interface{} {
	eface := reflectx.EfaceOf(&v)
	strMap := *(*map[string]interface{})(unsafe.Pointer(&eface.Word))
	return strMap
}

func getKeywordArgFunc(kwArgs interface{}) func(key string) (interface{}, bool) {
	if kwArgs == nil {
		return func(string) (interface{}, bool) { return nil, false }
	}
	kwTyp := reflect.TypeOf(kwArgs)
	if isStringInterfaceMap(kwTyp) {
		kwMap := castStringInterfaceMap(kwArgs)
		return func(key string) (interface{}, bool) {
			val, ok := kwMap[key]
			return val, ok
		}
	}
	if kwTyp.Kind() == reflect.Map && kwTyp.Key().Kind() == reflect.String {
		kwValue := reflect.ValueOf(kwArgs)
		return func(key string) (interface{}, bool) {
			val := kwValue.MapIndex(reflect.ValueOf(key))
			if val.IsValid() {
				return val.Interface(), true
			}
			return nil, false
		}
	}
	value := reflect.Indirect(reflect.ValueOf(kwArgs))
	if value.Kind() == reflect.Struct {
		return func(field string) (interface{}, bool) {
			x := value.FieldByName(field)
			if x.IsValid() {
				return x, true
			}
			return nil, false
		}
	}
	return func(string) (interface{}, bool) { return nil, false }
}
