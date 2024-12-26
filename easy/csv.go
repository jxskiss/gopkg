package easy

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"unsafe"

	"github.com/jxskiss/gopkg/v2/easy/ezmap"
	"github.com/jxskiss/gopkg/v2/utils/strutil"
)

// MarshalCSV marshal map[string]any records to CSV encoding.
// It uses the std encoding/csv.Writer with its default settings for csv encoding.
// If length of records is zero, it returns (nil, nil).
//
// Caller should guarantee that every record have same schema.
// The keys of the first item in records is used as the result CSV header,
// for the left items in records, if a key is missing, it is ignored,
// keys not present in the first item are simply ignored.
func MarshalCSV[T ~map[string]any](records []T) ([]byte, error) {
	if len(records) == 0 {
		return nil, nil
	}

	header := make([]string, 0, len(records[0]))
	for k := range records[0] {
		header = append(header, k)
	}
	sort.Strings(header)

	var err error
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if err = w.Write(header); err != nil {
		return nil, err
	}

	var strRecord []string
	for _, r := range records {
		strRecord = strRecord[:0]
		for _, k := range header {
			v, ok := r[k]
			if !ok {
				strRecord = append(strRecord, "")
				continue
			}
			var valueStr string
			kind := reflect.TypeOf(v).Kind()
			if kind == reflect.String {
				valueStr = castStr(v)
			} else if isSimpleType(kind) {
				valueStr = fmt.Sprint(v)
			} else {
				valueStr, err = toJSON(v)
				if err != nil {
					return nil, err
				}
			}
			strRecord = append(strRecord, valueStr)
		}
		if err = w.Write(strRecord); err != nil {
			return nil, err
		}
	}
	w.Flush()
	if err = w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// UnmarshalCVS parses CSV-encoded data to map[string]any records.
// It uses the std encoding/csv.Reader with its default settings for csv encoding.
// The first record parsed from the first row is treated as CSV header,
// and used as the result map keys.
func UnmarshalCVS(data []byte) ([]ezmap.Map, error) {
	return UnmarshalCSVWithSeparator(data, ',')
}

// UnmarshalCSVWithSeparator is same to [UnmarshalCVS],
// except that it allows caller to specify the separator.
func UnmarshalCSVWithSeparator(data []byte, sep rune) ([]ezmap.Map, error) {
	if sep != ',' && sep != ';' && sep != '\t' {
		return nil, fmt.Errorf("unsupported separator: %c", sep)
	}
	data = strutil.TrimBOM(data)
	csvReader := csv.NewReader(bytes.NewReader(data))
	csvReader.Comma = sep
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("csv.Reader.ReadAll: %w", err)
	}
	if len(records) <= 1 {
		return nil, nil
	}
	header, records := records[0], records[1:]
	for i, x := range header {
		for j := i + 1; j < len(header); j++ {
			if x == header[j] {
				return nil, fmt.Errorf("duplicate header: %s", x)
			}
		}
	}
	out := make([]ezmap.Map, 0, len(records))
	for _, record := range records {
		m := make(ezmap.Map, len(header))
		for i, x := range record {
			m[header[i]] = x
		}
		out = append(out, m)
	}
	return out, nil
}

func isSimpleType(kind reflect.Kind) bool {
	switch kind {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64,
		reflect.String:
		return true
	default:
		return false
	}
}

func toJSON(v any) (string, error) {
	buf, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("cannot marshal value of type %T to JSON: %w", v, err)
	}
	return string(buf), nil
}

func castStr(v any) string {
	// type eface struct { rtype unsafe.Pointer, word unsafe.Pointer }
	return *(*string)((*[2]unsafe.Pointer)(unsafe.Pointer(&v))[1])
}
