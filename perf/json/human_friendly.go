package json

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/jxskiss/gopkg/v2/internal/unsafeheader"
)

// HumanFriendly is a marshaler implementation which generates data
// that is friendly for human reading.
// Also, this config can encode data with `interface{}` as map keys,
// in contrast, the standard library fails in this case.
// This utility is not designed for performance sensitive use-case.
var HumanFriendly = humanFriendlyImpl{}

type humanFriendlyImpl struct{}

// float64With6Digits wraps float64 to marshal with exactly 6 decimal places.
type float64With6Digits float64

func (f float64With6Digits) MarshalJSON() ([]byte, error) {
	bs := fmt.Sprintf("%.6f", float64(f))
	if strings.IndexByte(bs, '.') >= 0 {
		n := len(bs)
		x := 0
		for i := n - 1; i >= 0; i-- {
			if bs[i] == '0' {
				x++
				continue
			}
			if bs[i] == '.' {
				x++
			}
			break
		}
		bs = bs[:n-x]
	}
	return unsafeheader.StringToBytes(bs), nil
}

// convertAnyKeyMap converts map[any]T to map[string]any recursively,
// sorting keys and converting float64 values to 6 decimal places.
func convertAnyKeyMap(v any) any {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Map:
		keys := rv.MapKeys()
		sort.Slice(keys, func(i, j int) bool {
			return keyString(keys[i]) < keyString(keys[j])
		})
		result := make(map[string]any, len(keys))
		for _, key := range keys {
			val := rv.MapIndex(key)
			result[keyString(key)] = convertAnyKeyMap(val.Interface())
		}
		return result
	case reflect.Slice, reflect.Array:
		result := make([]any, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			result[i] = convertAnyKeyMap(rv.Index(i).Interface())
		}
		return result
	case reflect.Float32, reflect.Float64:
		f := rv.Float()
		// Only wrap with 6 decimal places if the value has fractional part
		if f != float64(int64(f)) {
			return float64With6Digits(f)
		}
		return f
	case reflect.Ptr:
		if rv.IsNil() {
			return nil
		}
		return convertAnyKeyMap(rv.Elem().Interface())
	case reflect.Interface:
		if rv.IsNil() {
			return nil
		}
		return convertAnyKeyMap(rv.Interface())
	default:
		return v
	}
}

func keyString(key reflect.Value) string {
	switch key.Kind() {
	case reflect.String:
		return key.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(key.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(key.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%.6f", key.Float())
	case reflect.Bool:
		return strconv.FormatBool(key.Bool())
	default:
		return fmt.Sprintf("%v", key.Interface())
	}
}

func (humanFriendlyImpl) Marshal(v any) ([]byte, error) {
	converted := convertAnyKeyMap(v)
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(converted); err != nil {
		return nil, err
	}
	return bytes.TrimSpace(buf.Bytes()), nil
}

func (humanFriendlyImpl) MarshalToString(v any) (string, error) {
	buf, err := HumanFriendly.Marshal(v)
	if err != nil {
		return "", err
	}
	return unsafeheader.BytesToString(buf), nil
}

func (humanFriendlyImpl) MarshalIndent(v any, prefix, indent string) ([]byte, error) {
	converted := convertAnyKeyMap(v)
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent(prefix, indent)
	if err := enc.Encode(converted); err != nil {
		return nil, err
	}
	return bytes.TrimSpace(buf.Bytes()), nil
}

func (humanFriendlyImpl) MarshalIndentString(v any, prefix, indent string) (string, error) {
	buf, err := HumanFriendly.MarshalIndent(v, prefix, indent)
	if err != nil {
		return "", err
	}
	return unsafeheader.BytesToString(buf), nil
}

// NewEncoder ...
func (humanFriendlyImpl) NewEncoder(w io.Writer) *Encoder {
	return &Encoder{&hFriendlyEncoder{
		w:   w,
		enc: json.NewEncoder(w),
	}}
}

type hFriendlyEncoder struct {
	w      io.Writer
	enc    *json.Encoder
	prefix string
	indent string
}

func (h *hFriendlyEncoder) Encode(val any) error {
	converted := convertAnyKeyMap(val)
	h.enc.SetEscapeHTML(false)
	h.enc.SetIndent(h.prefix, h.indent)
	return h.enc.Encode(converted)
}

func (h *hFriendlyEncoder) SetEscapeHTML(on bool) {
	h.enc.SetEscapeHTML(on)
}

func (h *hFriendlyEncoder) SetIndent(prefix, indent string) {
	h.prefix = prefix
	h.indent = indent
}
