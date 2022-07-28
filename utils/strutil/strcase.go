package strutil

import (
	"unicode"
)

// ToSnakeCase convert the given string to snake case following the Golang format:
// acronyms are converted to lower-case and preceded by an underscore.
func ToSnakeCase(in string) string {
	runes := []rune(in)
	length := len(runes)

	var out []rune
	for i := 0; i < length; i++ {
		if i > 0 && unicode.IsUpper(runes[i]) &&
			((i+1 < length && unicode.IsLower(runes[i+1])) || unicode.IsLower(runes[i-1])) {
			out = append(out, '_')
		}
		out = append(out, unicode.ToLower(runes[i]))
	}

	return string(out)
}

// ToCamelCase converts the given string to CamelCase.
func ToCamelCase(in string) string {
	return string(toCamelCase([]rune(in)))
}

// ToLowerCamelCase converts the given string to lowerCamelCase.
func ToLowerCamelCase(in string) string {
	out := toCamelCase([]rune(in))
	length := len(out)
	for i := 0; i < length; i++ {
		isUpper := unicode.IsUpper(out[i])
		if isUpper && (i == 0 || i+1 == length || unicode.IsUpper(out[i+1])) {
			out[i] -= 'A' - 'a'
			continue
		}
		break
	}
	return string(out)
}

func toCamelCase(runes []rune) []rune {
	var out []rune
	var capNext = true
	for _, v := range runes {
		isUpper, isLower := unicode.IsUpper(v), unicode.IsLower(v)
		if capNext && isLower {
			v += 'A' - 'a'
		}
		if isUpper || isLower {
			out = append(out, v)
			capNext = false
		} else if unicode.IsNumber(v) {
			out = append(out, v)
			capNext = true
		} else {
			capNext = v == '_' || v == ' ' || v == '-' || v == '.'
		}
	}
	return out
}
