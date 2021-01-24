package strutil

import "unicode"

// Unicode byte order mark (BOM) constants. Reference:
// - https://en.wikipedia.org/wiki/Byte_order_mark
// - https://www.unicode.org/faq/utf_bom.html
const (
	BOM_UTF8               = "\xEF\xBB\xBF"
	BOM_UTF16_BigEndian    = "\xFE\xFF"
	BOM_UTF16_LittleEndian = "\xFF\xFE"
	BOM_UTF32_BigEndian    = "\x00\x00\xFE\xFF"
	BOM_UTF32_LittleEndian = "\xFF\xFE\x00\x00"
)

const (
	AlphaLower  = "abcdefghijklmnopqrstuvwxyz"
	AlphaUpper  = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	Letters     = AlphaLower + AlphaUpper
	Digits      = "0123456789"
	HexDigits   = "0123456789abcdefABCDEF"
	OctDigits   = "01234567"
	AlphaDigits = Digits + Letters
	Punctuation = "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"
	Whitespace  = " \t\n\r\x0b\x0c"
	Printable   = Digits + Letters + Punctuation + Whitespace

	PunctNoEscape = "!#$%&()*+,-./:;<=>?@[]^_{|}~" // without " ' ` \
)

// IsASCII returns true if the string is empty or all characters in the
// string are ASCII, false otherwise.
func IsASCII(str string) bool {
	for _, x := range str {
		if x > unicode.MaxASCII {
			return false
		}
	}
	return true
}

// IsASCIIDigit returns true if all characters in the string are in range 0-9
// and there is at least one character, false otherwise.
func IsASCIIDigit(str string) bool {
	if len(str) == 0 {
		return false
	}
	for _, x := range str {
		if !('0' <= x && x <= '9') {
			return false
		}
	}
	return true
}

// IsDigit returns true if all characters in the string are digits and
// there is at least one character, false otherwise
func IsDigit(str string) bool {
	if len(str) == 0 {
		return false
	}
	for _, x := range str {
		if !unicode.IsDigit(x) {
			return false
		}
	}
	return true
}

// IsLower returns true if all cased characters in the string are lowercase
// and there is at least one cased character, false otherwise.
func IsLower(str string) bool {
	hasCased := false
	for _, x := range str {
		isUpper, isLower := unicode.IsUpper(x), unicode.IsLower(x)
		if isUpper || isLower {
			hasCased = true
			if isUpper {
				return false
			}
		}
	}
	return hasCased
}

// IsUpper returns true if all cased characters in the string are uppercase
// and there is at least one cased character, false otherwise.
func IsUpper(str string) bool {
	hasCased := false
	for _, x := range str {
		isUpper, isLower := unicode.IsUpper(x), unicode.IsLower(x)
		if isUpper || isLower {
			hasCased = true
			if isLower {
				return false
			}
		}
	}
	return hasCased
}

// IsPrintable returns true if all characters in the string are printable
// or the string is empty, false otherwise.
func IsPrintable(str string) bool {
	for _, x := range str {
		if !unicode.IsPrint(x) {
			return false
		}
	}
	return true
}
