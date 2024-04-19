package easy

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"unicode/utf8"

	"github.com/mitchellh/mapstructure"
	"github.com/tidwall/gjson"

	"github.com/jxskiss/gopkg/v2/easy/ezmap"
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

type JSONPathMapping [][3]string

// ParseJSONRecordsWithMapping parses gjson.Result array to slice of
// map[string]any according to json path mapping.
func ParseJSONRecordsWithMapping(arr []gjson.Result, mapping JSONPathMapping) []ezmap.Map {
	out := make([]ezmap.Map, 0, len(arr))
	mapper := &jsonMapper{}
	for _, row := range arr {
		result := mapper.parseRecord(row, mapping)
		out = append(out, result)
	}
	return out
}

// ParseJSONRecords parses gjson.Result array to slice of *T
// according to json path mapping defined by struct tag "mapping".
//
// Note:
//
//  1. The type parameter T must be a struct
//  2. It does not support recursive types
//  3. It has very limited support for complex types of struct fields,
//     e.g. []any, []*Struct, []map[string]any,
//     map[string]*Struct, map[string]map[string]any
func ParseJSONRecords[T any](dst *[]*T, records []gjson.Result, opts ...JSONMapperOpt) error {
	var sample T
	if reflect.TypeOf(sample).Kind() != reflect.Struct {
		return errors.New("ParseJSONRecords: type T must be a struct")
	}
	mapper := &jsonMapper{
		opts: *(new(jsonMapperOptions).Apply(opts...)),
	}
	mapping, err := mapper.parseStructMapping(sample, nil)
	if err != nil {
		return err
	}
	out := make([]map[string]any, 0, len(records))
	for _, row := range records {
		result := mapper.parseRecord(row, mapping)
		out = append(out, result)
	}
	return mapstructure.Decode(out, &dst)
}

type jsonConvFunc func(j gjson.Result, path string) any

type jsonMapper struct {
	opts      jsonMapperOptions
	convFuncs map[[3]string]jsonConvFunc
}

func (p *jsonMapper) parseRecord(j gjson.Result, mapping JSONPathMapping) map[string]any {
	result := make(map[string]any)
	for _, x := range mapping {
		key, path := x[0], x[1]
		convFunc := p.getConvFunc(x)
		value := convFunc(j, path)
		result[key] = value
	}
	return result
}

func (p *jsonMapper) getConvFunc(m [3]string) jsonConvFunc {
	if f := p.convFuncs[m]; f != nil {
		return f
	}
	path, typ := m[1], m[2]
	f := p.newConvFunc(path, typ)
	if p.convFuncs == nil {
		p.convFuncs = make(map[[3]string]jsonConvFunc)
	}
	p.convFuncs[m] = f
	return f
}

func (p *jsonMapper) newConvFunc(path, typ string) func(j gjson.Result, path string) any {
	path = strings.TrimSpace(path)
	switch typ {
	case "", "str": // default "str"
		return func(j gjson.Result, path string) any {
			return j.Get(path).String()
		}
	case "bool":
		return func(j gjson.Result, path string) any {
			return j.Get(path).Bool()
		}
	case "int":
		return func(j gjson.Result, path string) any {
			return j.Get(path).Int()
		}
	case "uint":
		return func(j gjson.Result, path string) any {
			return j.Get(path).Uint()
		}
	case "float":
		return func(j gjson.Result, path string) any {
			return j.Get(path).Float()
		}
	case "time":
		return func(j gjson.Result, path string) any {
			return j.Get(path).Time()
		}
	case "map":
		var subMapping JSONPathMapping
		path, subMapping = p.parseSubMapping(path)
		if len(subMapping) > 0 {
			return func(j gjson.Result, _ string) any {
				out := make(map[string]any)
				for k, v := range j.Get(path).Map() {
					out[k] = p.parseRecord(v, subMapping)
				}
				return out
			}
		}
		return func(j gjson.Result, path string) any {
			return j.Get(path).Value()
		}
	case "array":
		var subMapping JSONPathMapping
		path, subMapping = p.parseSubMapping(path)
		if len(subMapping) > 0 {
			return func(j gjson.Result, _ string) any {
				out := make([]any, 0)
				for _, x := range j.Get(path).Array() {
					out = append(out, p.parseRecord(x, subMapping))
				}
				return out
			}
		}
		return func(j gjson.Result, path string) any {
			return j.Get(path).Value()
		}
	}
	// fallback any
	return func(j gjson.Result, path string) any {
		return j.Get(path).Value()
	}
}

func (p *jsonMapper) parseSubMapping(str string) (path string, subMapping JSONPathMapping) {
	str = strings.TrimSpace(str)
	nlIdx := strings.IndexByte(str, '\n')
	if nlIdx <= 0 {
		return "", nil
	}
	var mapping JSONPathMapping
	path = str[:nlIdx]
	err := json.Unmarshal([]byte(str[nlIdx+1:]), &mapping)
	if err != nil {
		panic(err)
	}
	return path, mapping
}

func (p *jsonMapper) parseStructMapping(sample any, seenTypes []reflect.Type) (JSONPathMapping, error) {
	var mapping JSONPathMapping
	structTyp := reflect.TypeOf(sample)
	for i := range seenTypes {
		if seenTypes[i] == structTyp {
			return nil, fmt.Errorf("recursive type not supported: %T", sample)
		}
	}
	seenTypes = append(seenTypes, structTyp)

	numField := structTyp.NumField()
	for i := 0; i < numField; i++ {
		field := structTyp.Field(i)
		jsonPath := field.Tag.Get("mapping")
		if jsonPath == "-" || !field.IsExported() {
			continue
		}
		if jsonPath == "" {
			jsonPath = field.Name
		}
		if dynPath := p.opts.DynamicMapping[jsonPath]; dynPath != "" {
			jsonPath = dynPath
		}
		kind := p.getKind(field.Type)
		var mappingType string
		switch kind {
		case reflect.String:
			mappingType = "str"
		case reflect.Bool:
			mappingType = "bool"
		case reflect.Int:
			mappingType = "int"
		case reflect.Uint:
			mappingType = "uint"
		case reflect.Float32:
			mappingType = "float"
		case reflect.Struct:
			if field.Type.String() == "time.Time" {
				mappingType = "time"
			} else {
				mappingType = "map"
				subMapping, err := p.parseStructMapping(reflect.New(field.Type).Elem().Interface(), seenTypes)
				if err != nil {
					return nil, err
				}
				tmp, err := json.Marshal(subMapping)
				if err != nil {
					panic(err)
				}
				jsonPath += "\n" + string(tmp)
			}
		case reflect.Array, reflect.Slice:
			elemTyp := field.Type.Elem()
			isSupported, isStruct := p.isSupportedElemType(elemTyp)
			if !isSupported {
				return nil, fmt.Errorf("unsupported array/slice element type: %v", elemTyp)
			}
			mappingType = "array"
			if isStruct {
				if elemTyp.Kind() == reflect.Pointer {
					elemTyp = elemTyp.Elem()
				}
				subMapping, err := p.parseStructMapping(reflect.New(elemTyp).Elem().Interface(), seenTypes)
				if err != nil {
					return nil, err
				}
				tmp, err := json.Marshal(subMapping)
				if err != nil {
					panic(err)
				}
				jsonPath += "\n" + string(tmp)
			}
		case reflect.Map:
			elemTyp := field.Type.Elem()
			isSupported, isStruct := p.isSupportedElemType(elemTyp)
			if !isSupported {
				return nil, fmt.Errorf("unsupported map element type: %v", elemTyp)
			}
			mappingType = "map"
			if isStruct {
				if elemTyp.Kind() == reflect.Pointer {
					elemTyp = elemTyp.Elem()
				}
				subMapping, err := p.parseStructMapping(reflect.New(elemTyp).Elem().Interface(), seenTypes)
				if err != nil {
					return nil, err
				}
				tmp, err := json.Marshal(subMapping)
				if err != nil {
					panic(err)
				}
				jsonPath += "\n" + string(tmp)
			}
		default:
			return nil, fmt.Errorf("unsupported field type: %v", field.Type)
		}
		mapping = append(mapping, [3]string{field.Name, jsonPath, mappingType})
	}
	return mapping, nil
}

var (
	anyTyp    = reflect.TypeOf((*any)(nil)).Elem()
	anyMapTyp = reflect.TypeOf((*map[string]any)(nil)).Elem()
)

func (p *jsonMapper) isSupportedElemType(typ reflect.Type) (isSupported, isStruct bool) {
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	kind := p.getKind(typ)
	if typ == anyTyp || typ == anyMapTyp ||
		kind == reflect.Bool ||
		kind == reflect.Int ||
		kind == reflect.Uint ||
		kind == reflect.Float32 ||
		kind == reflect.String ||
		kind == reflect.Struct {
		isSupported = true
	}
	if kind == reflect.Struct {
		isStruct = true
	}
	return
}

func (p *jsonMapper) getKind(typ reflect.Type) reflect.Kind {
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	kind := typ.Kind()
	switch {
	case kind >= reflect.Int && kind <= reflect.Int64:
		return reflect.Int
	case kind >= reflect.Uint && kind <= reflect.Uint64:
		return reflect.Uint
	case kind == reflect.Float32 || kind == reflect.Float64:
		return reflect.Float32
	default:
		return kind
	}
}

// JSONMapperOpt customizes the behavior of parsing JSON records.
type JSONMapperOpt struct {
	apply func(options *jsonMapperOptions)
}

type jsonMapperOptions struct {
	DynamicMapping map[string]string
}

func (p *jsonMapperOptions) Apply(opts ...JSONMapperOpt) *jsonMapperOptions {
	for _, opt := range opts {
		opt.apply(p)
	}
	return p
}

// WithDynamicJSONMapping specifies dynamic JSON path mapping to use,
// if a key specified by struct tag "mapping" is found in mapping,
// the JSON path expression is replaced by the value from mapping.
func WithDynamicJSONMapping(mapping map[string]string) JSONMapperOpt {
	return JSONMapperOpt{
		apply: func(options *jsonMapperOptions) {
			options.DynamicMapping = mapping
		},
	}
}
