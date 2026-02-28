package json

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"

	"github.com/jxskiss/gopkg/v2/internal/unsafeheader"
)

// HumanFriendly is a marshaler implementation which generates data
// that is friendly for human reading.
// This utility is not designed for performance sensitive use-case.
var HumanFriendly = humanFriendlyImpl{}

type humanFriendlyImpl struct{}

// convertAnyKeyMap converts the map with interface{} key type to a map with string key type.
func convertAnyKeyMap(v any) any {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Map {
		return v
	}
	rt := rv.Type()
	keyType := rt.Key()
	if keyType.Kind() != reflect.Interface {
		return v
	}
	strTyp := reflect.TypeOf("")
	newMap := reflect.MakeMapWithSize(reflect.MapOf(strTyp, rt.Elem()), rv.Len())
	for iter := rv.MapRange(); iter.Next(); {
		key, val := iter.Key(), iter.Value()
		newMap.SetMapIndex(reflect.ValueOf(keyString(key)), val)
	}
	return newMap.Interface()
}

func formatFloat6Digits(v float64) string {
	bs := fmt.Sprintf("%.6f", v)
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
	return bs
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
		return formatFloat6Digits(key.Float())
	case reflect.Bool:
		return strconv.FormatBool(key.Bool())
	default:
		return fmt.Sprintf("%v", key.Interface())
	}
}

func (humanFriendlyImpl) Marshal(v any) ([]byte, error) {
	v = convertAnyKeyMap(v)
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
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
	v = convertAnyKeyMap(v)
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent(prefix, indent)
	if err := enc.Encode(v); err != nil {
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
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return &Encoder{&hFriendlyEncoder{
		w:   w,
		enc: enc,
	}}
}

type hFriendlyEncoder struct {
	w      io.Writer
	enc    *json.Encoder
	prefix string
	indent string
}

func (h *hFriendlyEncoder) Encode(val any) error {
	h.enc.SetIndent(h.prefix, h.indent)
	val = convertAnyKeyMap(val)
	return h.enc.Encode(val)
}

func (h *hFriendlyEncoder) SetEscapeHTML(on bool) {
	h.enc.SetEscapeHTML(on)
}

func (h *hFriendlyEncoder) SetIndent(prefix, indent string) {
	h.prefix = prefix
	h.indent = indent
}
