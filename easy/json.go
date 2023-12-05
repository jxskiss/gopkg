package easy

import (
	"bytes"
	"fmt"
	"unicode/utf8"

	"github.com/jxskiss/gopkg/v2/internal/unsafeheader"
	"github.com/jxskiss/gopkg/v2/perf/json"
)

// JSON converts given object to a json string, it never returns error.
// The marshalling method used here does not escape HTML characters,
// and map keys are sorted, which helps human reading.
func JSON(v any) string {
	b, err := json.HumanFriendly.Marshal(v)
	if err != nil {
		return fmt.Sprintf("<error: %v>", err)
	}
	b = bytes.TrimSpace(b)
	return unsafeheader.BytesToString(b)
}

// LazyJSON returns a lazy object which wraps v, and it marshals v
// using JSON when it's String method is called.
// This helps to avoid unnecessary marshaling in some use case,
// such as leveled logging.
func LazyJSON(v any) fmt.Stringer {
	return lazyString{f: JSON, v: v}
}

// LazyFunc returns a lazy object which wraps v,
// which marshals v using f when it's String method is called.
// This helps to avoid unnecessary marshaling in some use case,
// such as leveled logging.
func LazyFunc(v any, f func(any) string) fmt.Stringer {
	return lazyString{f: f, v: v}
}

type lazyString struct {
	f func(any) string
	v any
}

func (x lazyString) String() string { return x.f(x.v) }

// Pretty converts given object to a pretty formatted json string.
// If the input is a json string, it will be formatted using json.Indent
// with four space characters as indent.
func Pretty(v any) string {
	return prettyIndent(v, "    ")
}

// Pretty2 is like Pretty, but it uses two space characters as indent,
// instead of four.
func Pretty2(v any) string {
	return prettyIndent(v, "  ")
}

func prettyIndent(v any, indent string) string {
	var src []byte
	switch v := v.(type) {
	case []byte:
		src = v
	case string:
		src = unsafeheader.StringToBytes(v)
	}
	if src != nil {
		if json.Valid(src) {
			buf := bytes.NewBuffer(nil)
			_ = json.Indent(buf, src, "", indent)
			return unsafeheader.BytesToString(buf.Bytes())
		}
		if utf8.Valid(src) {
			return string(src)
		}
		return fmt.Sprintf("<pretty: non-printable bytes of length %d>", len(src))
	}
	buf, err := json.HumanFriendly.MarshalIndent(v, "", indent)
	if err != nil {
		return fmt.Sprintf("<error: %v>", err)
	}
	buf = bytes.TrimSpace(buf)
	return unsafeheader.BytesToString(buf)
}
