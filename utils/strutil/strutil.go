package strutil

import (
	"unicode"
	"unicode/utf8"
)

// equalFoldRune compares a and b runes whether they fold equally.
//
// The code comes from strings.EqualFold, but shortened to only one rune.
func equalFoldRune(sr, tr rune) bool {
	if sr == tr {
		return true
	}
	// Make sr < tr to simplify what follows.
	if tr < sr {
		sr, tr = tr, sr
	}
	// Fast check for ASCII.
	if tr < utf8.RuneSelf && 'A' <= sr && sr <= 'Z' {
		// ASCII, and sr is upper case.  tr must be lower case.
		if tr == sr+'a'-'A' {
			return true
		}
		return false
	}

	// General case.  SimpleFold(x) returns the next equivalent rune > x
	// or wraps around to smaller values.
	r := unicode.SimpleFold(sr)
	for r != sr && r < tr {
		r = unicode.SimpleFold(r)
	}
	if r == tr {
		return true
	}
	return false
}

// HasPrefixFold is like strings.HasPrefix but uses Unicode case-folding,
// matching case insensitively.
func HasPrefixFold(s, prefix string) bool {
	if prefix == "" {
		return true
	}
	for _, pr := range prefix {
		if s == "" {
			return false
		}
		// step with s, too
		sr, size := utf8.DecodeRuneInString(s)
		if sr == utf8.RuneError {
			return false
		}
		s = s[size:]
		if !equalFoldRune(sr, pr) {
			return false
		}
	}
	return true
}

// HasSuffixFold is like strings.HasSuffix but uses Unicode case-folding,
// matching case insensitively.
func HasSuffixFold(s, suffix string) bool {
	if suffix == "" {
		return true
	}
	// count the runes and bytes in s, but only till rune count of suffix
	bo, so := len(s), len(suffix)
	for bo > 0 && so > 0 {
		r, size := utf8.DecodeLastRuneInString(s[:bo])
		if r == utf8.RuneError {
			return false
		}
		bo -= size

		sr, size := utf8.DecodeLastRuneInString(suffix[:so])
		if sr == utf8.RuneError {
			return false
		}
		so -= size

		if !equalFoldRune(r, sr) {
			return false
		}
	}
	return so == 0
}

// ContainsFold is like strings.Contains but uses Unicode case-folding.
func ContainsFold(s, substr string) bool {
	if substr == "" {
		return true
	}
	if s == "" {
		return false
	}
	firstRune := rune(substr[0])
	if firstRune >= utf8.RuneSelf {
		firstRune, _ = utf8.DecodeRuneInString(substr)
	}
	for i, rune := range s {
		if equalFoldRune(rune, firstRune) && HasPrefixFold(s[i:], substr) {
			return true
		}
	}
	return false
}

// Reverse returns a new string of the rune characters from the given string
// in reverse order.
func Reverse(s string) string {
	runes := []rune(s)
	length := len(runes)
	i, j := 0, length-1
	for i < j {
		runes[i], runes[j] = runes[j], runes[i]
		i++
		j--
	}
	return string(runes)
}
