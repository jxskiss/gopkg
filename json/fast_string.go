package json

import (
	"fmt"
	"unicode/utf16"
	"unicode/utf8"
)

const (
	comma     = ','
	colon     = ':'
	quotation = '"'
	leftWING  = '{'
	rightWING = '}'
	leftBRK   = '['
	rightBRK  = ']'
)

const (
	t1 = 0x00 // 0000 0000
	tx = 0x80 // 1000 0000
	t2 = 0xC0 // 1100 0000
	t3 = 0xE0 // 1110 0000
	t4 = 0xF0 // 1111 0000
	t5 = 0xF8 // 1111 1000

	maskx = 0x3F // 0011 1111
	mask2 = 0x1F // 0001 1111
	mask3 = 0x0F // 0000 1111
	mask4 = 0x07 // 0000 0111

	rune1Max = 1<<7 - 1
	rune2Max = 1<<11 - 1
	rune3Max = 1<<16 - 1

	surrogateMin = 0xD800
	surrogateMax = 0xDFFF

	maxRune   = '\U0010FFFF' // Maximum valid Unicode code point.
	runeError = '\uFFFD'     // the "error" Rune or "Unicode replacement character"

	hex = "0123456789abcdef"
)

var (
	nullJSON    = []byte("null")
	emptyObject = []byte("{}")
)

func appendString(buf []byte, s string) ([]byte, error) {
	valLen := len(s)
	buf = append(buf, quotation)
	// write string, the fast path, without utf8 and escape
	i := 0
	for ; i < valLen; i++ {
		c := s[i]
		if c < utf8.RuneSelf && htmlSafeSet[c] {
			buf = append(buf, c)
		} else {
			break
		}
	}
	if i == valLen {
		buf = append(buf, quotation)
		return buf, nil
	}
	return appendStringSlowPath(buf, i, s, valLen), nil
}

func appendStringSlowPath(buf []byte, i int, s string, valLen int) []byte {
	start := i
	for i < valLen {
		if b := s[i]; b < utf8.RuneSelf {
			if htmlSafeSet[b] {
				i++
				continue
			}
			if start < i {
				buf = append(buf, s[start:i]...)
			}
			switch b {
			case '\\', '"':
				buf = append(buf, '\\', b)
			case '\n':
				buf = append(buf, '\\', 'n')
			case '\r':
				buf = append(buf, '\\', 'r')
			case '\t':
				buf = append(buf, '\\', 't')
			default:
				// This encodes bytes < 0x20 except for \t, \n and \r.
				// If escapeHTML is set, it also escapes <, >, and &
				// because they can lead to security holes when
				// user-controlled strings are rendered into JSON
				// and served to some browsers.
				buf = append(buf, `\u00`...)
				buf = append(buf, hex[b>>4], hex[b&0xF])
			}
			i++
			start = i
			continue
		}
		c, size := utf8.DecodeRuneInString(s[i:])
		if c == utf8.RuneError && size == 1 {
			if start < i {
				buf = append(buf, s[start:i]...)
			}
			buf = append(buf, `\ufffd`...)
			i++
			start = i
			continue
		}
		// U+2028 is LINE SEPARATOR.
		// U+2029 is PARAGRAPH SEPARATOR.
		// They are both technically valid characters in JSON strings,
		// but don't work in JSONP, which has to be evaluated as JavaScript,
		// and can lead to security holes there. It is valid JSON to
		// escape them, so we do so unconditionally.
		// See http://timelessrepo.com/json-isnt-a-javascript-subset for discussion.
		if c == '\u2028' || c == '\u2029' {
			if start < i {
				buf = append(buf, s[start:i]...)
			}
			buf = append(buf, `\u202`...)
			buf = append(buf, hex[c&0xF])
			i += size
			start = i
			continue
		}
		i += size
	}
	if start < len(s) {
		buf = append(buf, s[start:]...)
	}
	buf = append(buf, quotation)
	return buf
}

// htmlSafeSet holds the value true if the ASCII character with the given
// array position can be safely represented inside a JSON string, embedded
// inside of HTML <script> tags, without any additional escaping.
//
// All values are true except for the ASCII control characters (0-31), the
// double quote ("), the backslash character ("\"), HTML opening and closing
// tags ("<" and ">"), and the ampersand ("&").
var htmlSafeSet = [utf8.RuneSelf]bool{
	' ':      true,
	'!':      true,
	'"':      false,
	'#':      true,
	'$':      true,
	'%':      true,
	'&':      false,
	'\'':     true,
	'(':      true,
	')':      true,
	'*':      true,
	'+':      true,
	',':      true,
	'-':      true,
	'.':      true,
	'/':      true,
	'0':      true,
	'1':      true,
	'2':      true,
	'3':      true,
	'4':      true,
	'5':      true,
	'6':      true,
	'7':      true,
	'8':      true,
	'9':      true,
	':':      true,
	';':      true,
	'<':      false,
	'=':      true,
	'>':      false,
	'?':      true,
	'@':      true,
	'A':      true,
	'B':      true,
	'C':      true,
	'D':      true,
	'E':      true,
	'F':      true,
	'G':      true,
	'H':      true,
	'I':      true,
	'J':      true,
	'K':      true,
	'L':      true,
	'M':      true,
	'N':      true,
	'O':      true,
	'P':      true,
	'Q':      true,
	'R':      true,
	'S':      true,
	'T':      true,
	'U':      true,
	'V':      true,
	'W':      true,
	'X':      true,
	'Y':      true,
	'Z':      true,
	'[':      true,
	'\\':     false,
	']':      true,
	'^':      true,
	'_':      true,
	'`':      true,
	'a':      true,
	'b':      true,
	'c':      true,
	'd':      true,
	'e':      true,
	'f':      true,
	'g':      true,
	'h':      true,
	'i':      true,
	'j':      true,
	'k':      true,
	'l':      true,
	'm':      true,
	'n':      true,
	'o':      true,
	'p':      true,
	'q':      true,
	'r':      true,
	's':      true,
	't':      true,
	'u':      true,
	'v':      true,
	'w':      true,
	'x':      true,
	'y':      true,
	'z':      true,
	'{':      true,
	'|':      true,
	'}':      true,
	'~':      true,
	'\u007f': true,
}

func nextToken(buf []byte, idx int, lastIdx int) (byte, int, error) {
	if lastIdx < idx {
		return 0, -1, fmt.Errorf("nextToken no more data")
	}
	var c byte
	for idx <= lastIdx {
		c = buf[idx]
		idx++
		switch c {
		case ' ', '\n', '\t', '\r':
			continue
		}
		return c, idx, nil
	}
	return c, idx, nil
}

func readByte(buf []byte, idx int, lastIdx int) (byte, int, error) {
	if lastIdx < idx {
		return 0, -1, fmt.Errorf("readByte no more data")
	}
	c := buf[idx]
	idx++
	return c, idx, nil
}

func checkNull(c byte, data []byte, idx int, lastIdx int) (int, bool) {
	if c == 'n' {
		ch, idx, _ := readByte(data, idx, lastIdx)
		if ch != 'u' {
			idx--
			return idx, false
		}
		ch, idx, _ = readByte(data, idx, lastIdx)
		if ch != 'l' {
			idx--
			return idx, false
		}
		ch, idx, _ = readByte(data, idx, lastIdx)
		if ch != 'l' {
			idx--
			return idx, false
		}
		return idx, true
	}
	return idx, false
}

func readU4(buf []byte, idx int, lastIdx int) (rune, int, error) {
	var err error
	var ret rune
	for i := 0; i < 4; i++ {
		var c byte
		c, idx, err = readByte(buf, idx, lastIdx)
		if err != nil {
			return ret, idx, err
		}
		if c >= '0' && c <= '9' {
			ret = ret*16 + rune(c-'0')
		} else if c >= 'a' && c <= 'f' {
			ret = ret*16 + rune(c-'a'+10)
		} else if c >= 'A' && c <= 'F' {
			ret = ret*16 + rune(c-'A'+10)
		} else {
			err = fmt.Errorf("readU4 expects 0~9 or a~f, but found %v", string([]byte{c}))
			return ret, idx, err
		}
	}
	return ret, idx, nil
}

func readString(buf []byte, idx int, lastIdx int) (string, int, error) {
	var err error
	var c byte
	var isNull bool
	c, idx, err = nextToken(buf, idx, lastIdx)
	if err != nil {
		return "", idx, err
	}
	var str []byte
	if c == '"' {
		start := idx
		var noESC = true
		for idx <= lastIdx {
			c, idx, err = readByte(buf, idx, lastIdx)
			if err != nil {
				return "", idx, err
			}
			switch c {
			case '"':
				if start < idx-1 {
					if noESC {
						str = buf[start : idx-1]
					} else {
						str = append(str, buf[start:idx-1]...)
					}
				}
				return b2s(str), idx, nil
			case '\\':
				if start < idx-1 {
					if noESC {
						str = buf[start : idx-1]
					} else {
						str = append(str, buf[start:idx-1]...)
					}
				}
				c, idx, err = readByte(buf, idx, lastIdx)
				if err != nil {
					return "", idx, err
				}
				str, idx, err = readEscapedChar(c, buf, idx, str, lastIdx)
				if err != nil {
					return "", idx, err
				}
				start = idx
				noESC = false
			}
		}
	} else if idx, isNull = checkNull(c, buf, idx, lastIdx); isNull {
		return "", idx, nil
	}
	err = fmt.Errorf("readString expects '\"' or n, but found %s", string(c))
	return b2s(str), idx, err
}

func readEscapedChar(c byte, buf []byte, idx int, str []byte, lastIdx int) ([]byte, int, error) {
	var err error
	switch c {
	case 'u':
		var r rune
		r, idx, err = readU4(buf, idx, lastIdx)
		if err != nil {
			return str, idx, err
		}
		if utf16.IsSurrogate(r) {
			c, idx, err = readByte(buf, idx, lastIdx)
			if err != nil {
				return str, idx, err
			}
			if c != '\\' {
				idx--
				str = appendRune(str, r)
				return str, idx, nil
			}
			c, idx, err = readByte(buf, idx, lastIdx)
			if err != nil {
				return str, idx, err
			}
			if c != 'u' {
				str = appendRune(str, r)
				return readEscapedChar(c, buf, idx, str, lastIdx)
			}
			var r2 rune
			r2, idx, err = readU4(buf, idx, lastIdx)
			if err != nil {
				return str, idx, err
			}
			combined := utf16.DecodeRune(r, r2)
			if combined == '\uFFFD' {
				str = appendRune(str, r)
				str = appendRune(str, r2)
			} else {
				str = appendRune(str, combined)
			}
		} else {
			str = appendRune(str, r)
		}
	case '"':
		str = append(str, '"')
	case '\\':
		str = append(str, '\\')
	case '/':
		str = append(str, '/')
	case 'b':
		str = append(str, '\b')
	case 'f':
		str = append(str, '\f')
	case 'n':
		str = append(str, '\n')
	case 'r':
		str = append(str, '\r')
	case 't':
		str = append(str, '\t')
	default:
		err = fmt.Errorf("readEscapedChar found invalid escape char after \\")
		return str, idx, err
	}
	return str, idx, nil
}

func appendRune(p []byte, r rune) []byte {
	// Negative values are erroneous. Making it unsigned addresses the problem.
	switch i := uint32(r); {
	case i <= rune1Max:
		p = append(p, byte(r))
		return p
	case i <= rune2Max:
		p = append(p, t2|byte(r>>6))
		p = append(p, tx|byte(r)&maskx)
		return p
	case i > maxRune, surrogateMin <= i && i <= surrogateMax:
		r = runeError
		fallthrough
	case i <= rune3Max:
		p = append(p, t3|byte(r>>12))
		p = append(p, tx|byte(r>>6)&maskx)
		p = append(p, tx|byte(r)&maskx)
		return p
	default:
		p = append(p, t4|byte(r>>18))
		p = append(p, tx|byte(r>>12)&maskx)
		p = append(p, tx|byte(r>>6)&maskx)
		p = append(p, tx|byte(r)&maskx)
		return p
	}
}