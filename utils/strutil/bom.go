package strutil

import (
	"bufio"
	"io"
)

// Unicode byte order mark (BOM) constants.
//
// Reference:
// https://en.wikipedia.org/wiki/Byte_order_mark
// and https://www.unicode.org/faq/utf_bom.html
//
//nolint:all
const (
	BOM_UTF8               = "\xEF\xBB\xBF"
	BOM_UTF16_BigEndian    = "\xFE\xFF"
	BOM_UTF16_LittleEndian = "\xFF\xFE"
	BOM_UTF32_BigEndian    = "\x00\x00\xFE\xFF"
	BOM_UTF32_LittleEndian = "\xFF\xFE\x00\x00"
)

// DetectBOM detects BOM prefix from a byte slice.
func DetectBOM(b []byte) (bom string) {
	if len(b) >= 4 {
		first4 := string(b[:4])
		if first4 == BOM_UTF32_BigEndian || first4 == BOM_UTF32_LittleEndian {
			return first4
		}
	}
	if len(b) >= 3 {
		first3 := string(b[:3])
		if first3 == BOM_UTF8 {
			return first3
		}
	}
	if len(b) >= 2 {
		first2 := string(b[:2])
		if first2 == BOM_UTF16_BigEndian || first2 == BOM_UTF16_LittleEndian {
			return first2
		}
	}
	return ""
}

// TrimBOM detects and trims BOM prefix from a byte slice, the returned
// byte slice shares the same underlying memory with the given slice.
func TrimBOM(b []byte) []byte {
	bom := DetectBOM(b)
	if bom == "" {
		return b
	}
	return b[len(bom):]
}

// SkipBOMReader detects and skips BOM prefix from the given io.Reader.
// It returns a *bufio.Reader.
func SkipBOMReader(rd io.Reader) io.Reader {
	buf := bufio.NewReader(rd)
	first, err := buf.Peek(4)
	if err != nil {
		first, err = buf.Peek(3)
		if err != nil {
			first, err = buf.Peek(2)
			if err != nil { // not enough data
				return buf
			}
		}
	}
	bom := DetectBOM(first)
	_, _ = buf.Discard(len(bom))
	return buf
}
