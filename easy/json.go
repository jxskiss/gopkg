package easy

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"
	"unsafe"

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
// using JSON when its String method is called.
// This helps to avoid unnecessary marshaling in some use case,
// such as leveled logging.
func LazyJSON(v any) fmt.Stringer {
	return lazyString{f: JSON, v: v}
}

// LazyFunc returns a lazy object which wraps v,
// which marshals v using f when its String method is called.
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

// LazyFunc0 returns a lazy fmt.Stringer which calls f when
// its String method is called.
func LazyFunc0(f func() string) fmt.Stringer {
	return lazyString0{f: f}
}

type lazyString0 struct {
	f func() string
}

func (x lazyString0) String() string { return x.f() }

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

// ParseJSONToMaps parses gjson.Result array to slice of
// map[string]any according to json path mapping.
func ParseJSONToMaps(arr []gjson.Result, mapping JSONPathMapping) []ezmap.Map {
	out := make([]ezmap.Map, 0, len(arr))
	mapper := &jsonMapper{}
	convFuncs := mapper.getConvFuncs(mapping)
	for _, row := range arr {
		result := mapper.parseRecord(row, mapping, convFuncs)
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
//  2. It has very limited support for complex types of struct fields,
//     e.g. []any, []*Struct, []map[string]any,
//     map[string]any, map[string]*Struct, map[string]map[string]any
func ParseJSONRecords[T any](records *[]*T, arr []gjson.Result, opts ...JSONMapperOpt) error {
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
	convFuncs := mapper.getConvFuncs(mapping)
	out := make([]map[string]any, 0, len(arr))
	for _, row := range arr {
		result := mapper.parseRecord(row, mapping, convFuncs)
		out = append(out, result)
	}
	return mapstructure.Decode(out, records)
}

type jsonConvFunc func(j gjson.Result, path string) any

type jsonMapper struct {
	opts          jsonMapperOptions
	convFuncs     map[[3]string]jsonConvFunc
	structMapping map[string]JSONPathMapping
}

func (p *jsonMapper) parseRecord(j gjson.Result, mapping JSONPathMapping, convFuncs []jsonConvFunc) map[string]any {
	if !j.Exists() {
		return nil
	}
	result := make(map[string]any)
	for i, x := range mapping {
		key, path := x[0], x[1]
		value := convFuncs[i](j, path)
		result[key] = value
	}
	return result
}

func (p *jsonMapper) getConvFuncs(mapping JSONPathMapping) []jsonConvFunc {
	funcs := make([]jsonConvFunc, len(mapping))
	for i, x := range mapping {
		f, exists := p.convFuncs[x]
		if !exists {
			path, typ := x[1], x[2]
			f = p.newConvFunc(path, typ)
			if p.convFuncs == nil {
				p.convFuncs = make(map[[3]string]jsonConvFunc)
			}
			p.convFuncs[x] = f
		}
		funcs[i] = f
	}
	return funcs
}

func (p *jsonMapper) newConvFunc(path, typ string) jsonConvFunc {
	path = strings.TrimSpace(path)
	switch typ {
	case "", "str", "string": // default "str"
		return func(j gjson.Result, path string) any {
			return j.Get(path).String()
		}
	case "bool":
		return func(j gjson.Result, path string) any {
			return j.Get(path).Bool()
		}
	case "int", "int8", "int16", "int32", "int64":
		return func(j gjson.Result, path string) any {
			return j.Get(path).Int()
		}
	case "uint", "uint8", "uint16", "uint32", "uint64":
		return func(j gjson.Result, path string) any {
			return j.Get(path).Uint()
		}
	case "float", "float32", "float64":
		return func(j gjson.Result, path string) any {
			return j.Get(path).Float()
		}
	case "time":
		return func(j gjson.Result, path string) any {
			return j.Get(path).Time()
		}
	case "struct":
		var (
			subMapping   JSONPathMapping
			subConvFuncs []jsonConvFunc
		)
		path, subMapping = p.parseSubMapping(path)
		if len(subMapping) > 0 {
			return func(j gjson.Result, _ string) any {
				if !j.Exists() {
					return nil
				}
				if subConvFuncs == nil {
					subConvFuncs = p.getConvFuncs(subMapping)
				}
				return p.parseRecord(j.Get(path), subMapping, subConvFuncs)
			}
		}
	case "map":
		var (
			subMapping   JSONPathMapping
			subConvFuncs []jsonConvFunc
		)
		path, subMapping = p.parseSubMapping(path)
		if len(subMapping) > 0 {
			return func(j gjson.Result, _ string) any {
				j = j.Get(path)
				if !j.Exists() {
					return nil
				}
				if subConvFuncs == nil {
					subConvFuncs = p.getConvFuncs(subMapping)
				}
				out := make(map[string]any)
				for k, v := range j.Map() {
					out[k] = p.parseRecord(v, subMapping, subConvFuncs)
				}
				return out
			}
		}
	case "arr", "array", "slice":
		var (
			subMapping   JSONPathMapping
			subConvFuncs []jsonConvFunc
		)
		path, subMapping = p.parseSubMapping(path)
		if len(subMapping) > 0 {
			return func(j gjson.Result, _ string) any {
				j = j.Get(path)
				if !j.Exists() {
					return nil
				}
				if subConvFuncs == nil {
					subConvFuncs = p.getConvFuncs(subMapping)
				}
				out := make([]any, 0)
				for _, x := range j.Array() {
					out = append(out, p.parseRecord(x, subMapping, subConvFuncs))
				}
				return out
			}
		}
	}
	// fallback any
	return func(j gjson.Result, path string) any {
		j = j.Get(path)
		if !j.Exists() {
			return nil
		}
		return j.Value()
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
	subPath := str[nlIdx+1:]
	if p.isStructMappingKey(subPath) {
		mapping = p.structMapping[subPath]
	} else {
		err := json.Unmarshal([]byte(subPath), &mapping)
		if err != nil {
			panic(err)
		}
	}
	return path, mapping
}

func (p *jsonMapper) isStructMappingKey(s string) bool {
	return strings.HasPrefix(s, "\t\tSTRUCT\t")
}

func (p *jsonMapper) getStructMappingKey(typ reflect.Type) string {
	// type iface { tab  *itab, data unsafe.Pointer }
	typeptr := (*(*[2]uintptr)(unsafe.Pointer(&typ)))[1]
	return "\t\tSTRUCT\t" + strconv.FormatUint(uint64(typeptr), 32)
}

func (p *jsonMapper) parseStructMapping(sample any, seenTypes []reflect.Type) (mapping JSONPathMapping, err error) {
	structTyp := reflect.TypeOf(sample)
	seenTypes = append(seenTypes, structTyp)
	defer func() {
		if err == nil {
			key := p.getStructMappingKey(structTyp)
			if p.structMapping == nil {
				p.structMapping = make(map[string]JSONPathMapping)
			}
			p.structMapping[key] = mapping
		}
	}()

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
		_fieldTyp := field.Type
		if _fieldTyp.Kind() == reflect.Pointer {
			_fieldTyp = _fieldTyp.Elem()
		}
		kind := p.getKind(_fieldTyp)
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
			if _fieldTyp.String() == "time.Time" {
				mappingType = "time"
			} else {
				mappingType = "struct"
				subStructKey := p.getStructMappingKey(_fieldTyp)
				if _, ok := p.structMapping[subStructKey]; !ok &&
					!isSeenType(seenTypes, _fieldTyp) {
					_, err := p.parseStructMapping(reflect.New(_fieldTyp).Elem().Interface(), seenTypes)
					if err != nil {
						return nil, err
					}
				}
				jsonPath += "\n" + subStructKey
			}
		case reflect.Array, reflect.Slice:
			elemTyp := field.Type.Elem()
			if elemTyp.Kind() == reflect.Pointer {
				elemTyp = elemTyp.Elem()
			}
			isSupported, isStruct := p.isSupportedElemType(elemTyp)
			if !isSupported {
				return nil, fmt.Errorf("unsupported array/slice element type: %v", field.Type)
			}
			mappingType = "array"
			if isStruct {
				if elemTyp.Kind() == reflect.Pointer {
					elemTyp = elemTyp.Elem()
				}
				subStructKey := p.getStructMappingKey(elemTyp)
				if _, ok := p.structMapping[subStructKey]; !ok &&
					!isSeenType(seenTypes, elemTyp) {
					_, err := p.parseStructMapping(reflect.New(elemTyp).Elem().Interface(), seenTypes)
					if err != nil {
						return nil, err
					}
				}
				jsonPath += "\n" + subStructKey
			}
		case reflect.Map:
			keyType := field.Type.Key()
			if keyType.Kind() != reflect.String {
				return nil, fmt.Errorf("unsupported map key type: %v", field.Type)
			}
			elemTyp := field.Type.Elem()
			if elemTyp.Kind() == reflect.Pointer {
				elemTyp = elemTyp.Elem()
			}
			isSupported, isStruct := p.isSupportedElemType(elemTyp)
			if !isSupported {
				return nil, fmt.Errorf("unsupported map element type: %v", field.Type)
			}
			mappingType = "map"
			if isStruct {
				if elemTyp.Kind() == reflect.Pointer {
					elemTyp = elemTyp.Elem()
				}
				subStructKey := p.getStructMappingKey(elemTyp)
				if _, ok := p.structMapping[subStructKey]; !ok &&
					!isSeenType(seenTypes, elemTyp) {
					_, err := p.parseStructMapping(reflect.New(elemTyp).Elem().Interface(), seenTypes)
					if err != nil {
						return nil, err
					}
				}
				jsonPath += "\n" + subStructKey
			}
		default:
			return nil, fmt.Errorf("unsupported field type: %v", field.Type)
		}
		mapping = append(mapping, [3]string{field.Name, jsonPath, mappingType})
	}
	return mapping, nil
}

func isSeenType(stack []reflect.Type, typ reflect.Type) bool {
	for i := range stack {
		if stack[i] == typ {
			return true
		}
	}
	return false
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
