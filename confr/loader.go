// Package confr provides a simple but yet powerful configuration loader.
//
// Features:
//
// 1. Load from command line flags defined by field tag `flag`;
//
// 2. Load by custom loader function for fields which have a `custom` tag,
// this is useful where you may have configuration values stored in a
// centric repository or a remote config center;
//
// 3. Load from environment variables by explicitly defined `env` tag or
// auto-generated names implicitly;
//
// 4. Load from multiple configuration fields with priority and overriding;
//
// 5. Set default values by field tag `default` if a configuration field
// is not given by any of the higher priority source;
//
// 6. Minimal dependency;
//
// You may check Config and Loader for more details.
package confr

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path"
	"reflect"
	"strings"
	"time"
	"unicode"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cast"
	"gopkg.in/yaml.v3"

	"github.com/jxskiss/gopkg/v2/internal"
)

const DefaultEnvPrefix = "Confr"

const (
	ConfrTag        = "confr"
	CustomTag       = "custom"
	DefaultValueTag = "default"
	EnvTag          = "env"
	FlagTag         = "flag"
)

// Config provides options to configure the behavior of Loader.
type Config struct {

	// LogFunc specifies a log function to use instead of [log.Printf].
	LogFunc func(format string, v ...any)

	// Verbose tells the loader to output verbose logging messages.
	Verbose bool

	// DisallowUnknownFields causes the loader to return an error when
	// the configuration files contain object keys which do not match
	// the given destination struct.
	DisallowUnknownFields bool

	// EnableImplicitEnv enables the loader checking auto-generated names
	// to find environment variables.
	// The default is false, which means the loader will only check `env`
	// tag, won't check auto-generated names.
	EnableImplicitEnv bool

	// EnvPrefix is used to prefix the auto-generated names to find
	// environment variables. The default value is "Confr".
	EnvPrefix string

	// CustomLoader optionally loads fields which have a `custom` tag,
	// the field's type and the tag value will be passed to the custom loader.
	CustomLoader func(typ reflect.Type, tag string) (any, error)

	// UnmarshalFunc optionally specifies an unmarshal function
	// to use instead of the default function.
	UnmarshalFunc func(data []byte, v any, disallowUnknownFields bool) error

	// FlagSet optionally specifies a flag set to lookup flag value
	// for fields which have a `flag` tag. The tag value should be the
	// flag name to lookup for.
	FlagSet *flag.FlagSet
}

// Loader is used to load configuration from files (JSON/TOML/YAML),
// environment variables, command line flags, or by custom loader function.
//
// The priority in descending order is:
//
// 1. command line flag defined by field tag `flag`;
//
// 2. custom loader function defined by field tag `custom`;
//
// 3. environment variables;
//
// 4. config files, if multiple files are given to Load, files appeared
// first takes higher priority, if a config field appears in more
// than one files, only the first has effect.
//
// 5. default values defined by field tag `default`;
type Loader struct {
	*Config
}

// New creates a new Loader.
func New(config *Config) *Loader {
	if config == nil {
		config = &Config{}
	}

	return &Loader{Config: config}
}

// Load creates a Loader with nil config and loads configuration to dst,
// it is a shortcut for New(nil).Load(dst, files...).
func Load(dst any, files ...string) error {
	return New(nil).Load(dst, files...)
}

// Load loads configuration to dst using the Loader's Config
// and the given configuration files.
//
// See Loader and Config for detailed document.
func (p *Loader) Load(dst any, files ...string) error {
	return p.load(dst, files...)
}

func (p *Loader) load(dst any, files ...string) error {
	dstTyp := reflect.TypeOf(dst)
	if dstTyp.Kind() != reflect.Ptr {
		return errors.New("invalid destination, must be a pointer")
	}

	dstElemTyp := dstTyp.Elem()
	isStruct := dstElemTyp.Kind() == reflect.Struct
	isSlice := dstElemTyp.Kind() == reflect.Slice
	if !isStruct && !isSlice {
		return errors.New("invalid destination, must be a struct pointer or slice pointer")
	}

	if err := p.loadFiles(dst, files...); err != nil {
		return err
	}
	if err := p.processEnv(dst, ""); err != nil {
		return err
	}
	if err := p.processCustom(dst); err != nil {
		return err
	}
	if err := p.processDefaults(dst); err != nil {
		return err
	}
	if err := p.processFlags(dst); err != nil {
		return err
	}
	return nil
}

func (p *Loader) getLogFunc() func(format string, v ...any) {
	if p.LogFunc != nil {
		return p.LogFunc
	}
	return internal.DefaultLoggerInfof
}

func (p *Loader) loadFiles(config any, files ...string) error {
	for i := len(files) - 1; i >= 0; i-- {
		file := files[i]
		err := p.processFile(config, file)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Loader) processFile(config any, file string) error {
	if info, err := os.Stat(file); err != nil || !info.Mode().IsRegular() {
		return fmt.Errorf("invalid configuration file: %s", file)
	}

	p.getLogFunc()("loading configuration from file: %v", file)

	// Check if the target type is a slice and the file format is TOML
	configType := reflect.TypeOf(config)
	if configType.Kind() == reflect.Ptr {
		configType = configType.Elem()
	}
	extname := path.Ext(file)
	if configType.Kind() == reflect.Slice && strings.ToLower(extname) == ".toml" {
		return fmt.Errorf("toml format does not support top-level array, cannot load slice configuration from TOML file: %s", file)
	}

	var unmarshalFunc = p.UnmarshalFunc
	if unmarshalFunc == nil {
		switch strings.ToLower(extname) {
		case ".json":
			unmarshalFunc = unmarshalJSON
		case ".yaml", ".yml":
			unmarshalFunc = unmarshalYAML
		case ".toml":
			unmarshalFunc = unmarshalTOML
		default:
			return fmt.Errorf("unsupported file type: %v", extname)
		}
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("cannot read file %s: %w", file, err)
	}
	err = unmarshalFunc(data, config, p.DisallowUnknownFields)
	if err != nil {
		return fmt.Errorf("cannot unmarshal file %s: %w", file, err)
	}
	return nil
}

func unmarshalJSON(data []byte, v any, disallowUnknownFields bool) error {
	if disallowUnknownFields {
		dec := json.NewDecoder(bytes.NewReader(data))
		dec.DisallowUnknownFields()
		return dec.Decode(v)
	}
	return json.Unmarshal(data, v)
}

func unmarshalYAML(data []byte, v any, disallowUnknownFields bool) error {
	dec := yaml.NewDecoder(bytes.NewReader(data))
	if disallowUnknownFields {
		dec.KnownFields(true)
	}
	return dec.Decode(v)
}

func unmarshalTOML(data []byte, v any, disallowUnknownFields bool) error {
	targetType := reflect.TypeOf(v)
	if targetType.Kind() == reflect.Ptr {
		targetType = targetType.Elem()
	}

	if targetType.Kind() == reflect.Slice {
		// Try to decode the entire document to a generic interface
		var temp any
		_, err := toml.Decode(string(data), &temp)
		if err != nil {
			return err
		}

		// Check if the decoded result is a map (containing an array)
		if tempMap, ok := temp.(map[string]any); ok {
			// Get the target slice element type name
			elemTypeName := targetType.Elem().Name()

			// Find the array with the corresponding name
			if arr, found := tempMap[elemTypeName]; found {
				if slice, ok := arr.([]any); ok {
					// Convert the slice to the target type
					result := reflect.New(targetType).Elem() // Create a zero-value slice of the target type

					for _, item := range slice {
						// Convert each item to the target element type
						itemBytes, err := json.Marshal(item)
						if err != nil {
							return err
						}

						newItem := reflect.New(targetType.Elem()).Interface()
						err = json.Unmarshal(itemBytes, newItem)
						if err != nil {
							return err
						}

						// If newItem is a pointer type (pointing to a struct), take its element
						if reflect.TypeOf(newItem).Elem().Kind() == reflect.Struct {
							result = reflect.Append(result, reflect.ValueOf(newItem).Elem())
						} else {
							result = reflect.Append(result, reflect.ValueOf(newItem))
						}
					}

					// Set the target value
					reflect.ValueOf(v).Elem().Set(result)
					return nil
				}
			}
			// If no array with the specified name is found, try using the first array
			for _, arr := range tempMap {
				if slice, ok := arr.([]any); ok {
					// Convert the slice to the target type
					result := reflect.New(targetType).Elem() // Create a zero-value slice of the target type

					for _, item := range slice {
						// Convert each item to the target element type
						itemBytes, err := json.Marshal(item)
						if err != nil {
							return err
						}

						newItem := reflect.New(targetType.Elem()).Interface()
						err = json.Unmarshal(itemBytes, newItem)
						if err != nil {
							return err
						}

						// If newItem is a pointer type (pointing to a struct), take its element
						if reflect.TypeOf(newItem).Elem().Kind() == reflect.Struct {
							result = reflect.Append(result, reflect.ValueOf(newItem).Elem())
						} else {
							result = reflect.Append(result, reflect.ValueOf(newItem))
						}
					}

					// Set the target value
					reflect.ValueOf(v).Elem().Set(result)
					return nil
				}
			}
		} else if arr, ok := temp.([]any); ok { // If the decoded result is itself a slice (unlikely in TOML)
			// Create a new instance of the target slice type
			resultSlice := reflect.MakeSlice(targetType, 0, len(arr))

			// Convert each element to the target type
			for _, item := range arr {
				itemBytes, err := json.Marshal(item) // Use JSON as intermediate format
				if err != nil {
					return err
				}

				newItem := reflect.New(targetType.Elem()).Interface()
				err = json.Unmarshal(itemBytes, newItem)
				if err != nil {
					return err
				}

				// If newItem is a pointer, get the value it points to
				if reflect.TypeOf(newItem).Elem().Kind() == reflect.Struct {
					resultSlice = reflect.Append(resultSlice, reflect.ValueOf(newItem).Elem())
				} else {
					resultSlice = reflect.Append(resultSlice, reflect.ValueOf(newItem))
				}
			}

			// Set the target value
			reflect.ValueOf(v).Elem().Set(resultSlice)
			return nil
		}
	}

	// Default behavior
	meta, err := toml.Decode(string(data), v)
	if err == nil && len(meta.Undecoded()) > 0 && disallowUnknownFields {
		return fmt.Errorf("toml: unknown fields %v", meta.Undecoded())
	}
	return err
}

func (p *Loader) processDefaults(config any) error {
	configVal := reflect.Indirect(reflect.ValueOf(config))
	configTyp := configVal.Type()

	switch configTyp.Kind() {
	case reflect.Struct:
		for i := 0; i < configTyp.NumField(); i++ {
			field := configTyp.Field(i)
			fieldVal := configVal.Field(i)
			if !fieldVal.CanAddr() || !fieldVal.CanInterface() {
				continue
			}
			if field.Tag.Get(ConfrTag) == "-" {
				continue
			}

			defaultValue := field.Tag.Get(DefaultValueTag)
			if defaultValue != "" {
				if p.Verbose {
					p.getLogFunc()("processing default value for field %s.%s", configTyp.Name(), field.Name)
				}

				isBlank := reflect.DeepEqual(fieldVal.Interface(), reflect.Zero(field.Type).Interface())
				if isBlank {
					err := assignFieldValue(fieldVal, defaultValue)
					if err != nil {
						return fmt.Errorf("cannot assign default value to field %s.%s: %w", configTyp.Name(), field.Name, err)
					}
				}
			}

			fieldVal = reflect.Indirect(fieldVal)
			switch fieldVal.Kind() {
			case reflect.Struct:
				if err := p.processDefaults(fieldVal.Addr().Interface()); err != nil {
					return err
				}
			case reflect.Slice:
				// Recursively process struct elements in the slice
				for i := 0; i < fieldVal.Len(); i++ {
					elemVal := reflect.Indirect(fieldVal.Index(i))
					if elemVal.Kind() == reflect.Struct {
						if err := p.processDefaults(elemVal.Addr().Interface()); err != nil {
							return err
						}
					}
				}
			}
		}
	case reflect.Slice:
		// For slice types, recursively process each element
		for i := 0; i < configVal.Len(); i++ {
			elemVal := reflect.Indirect(configVal.Index(i))
			if elemVal.Kind() == reflect.Struct {
				if err := p.processDefaults(elemVal.Addr().Interface()); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (p *Loader) processFlags(config any) error {
	if p.FlagSet == nil {
		return nil
	}
	if !p.FlagSet.Parsed() {
		return errors.New("flag set is not parsed")
	}

	fs := p.FlagSet
	configVal := reflect.Indirect(reflect.ValueOf(config))
	configTyp := configVal.Type()

	switch configTyp.Kind() {
	case reflect.Struct:
		for i := 0; i < configTyp.NumField(); i++ {
			field := configTyp.Field(i)
			fieldVal := configVal.Field(i)
			if !fieldVal.CanAddr() || !fieldVal.CanInterface() {
				continue
			}
			if field.Tag.Get(ConfrTag) == "-" {
				continue
			}

			flagName := field.Tag.Get(FlagTag)
			if flagName != "" && flagName != "-" {
				if p.Verbose {
					p.getLogFunc()("processing flag for field %s.%s", configTyp.Name(), field.Name)
				}

				if flagVal, isSet := lookupFlag(fs, flagName); flagVal != nil {
					err := assignFlagValue(fieldVal, flagVal, isSet)
					if err != nil {
						return fmt.Errorf("cannot assign flag value to field %s.%s: %w", configTyp.Name(), field.Name, err)
					}
				}
			}

			fieldVal = reflect.Indirect(fieldVal)
			switch fieldVal.Kind() {
			case reflect.Struct:
				if err := p.processFlags(fieldVal.Addr().Interface()); err != nil {
					return err
				}
			case reflect.Slice:
				// Recursively process struct elements in the slice
				for i := 0; i < fieldVal.Len(); i++ {
					elemVal := reflect.Indirect(fieldVal.Index(i))
					if elemVal.Kind() == reflect.Struct {
						if err := p.processFlags(elemVal.Addr().Interface()); err != nil {
							return err
						}
					}
				}
			}
		}
	case reflect.Slice:
		// For slice types, recursively process each element
		for i := 0; i < configVal.Len(); i++ {
			elemVal := reflect.Indirect(configVal.Index(i))
			if elemVal.Kind() == reflect.Struct {
				if err := p.processFlags(elemVal.Addr().Interface()); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// lookupFlag returns a flag and tells whether the flag is set.
func lookupFlag(fs *flag.FlagSet, name string) (out *flag.Flag, isSet bool) {
	fs.Visit(func(f *flag.Flag) {
		if f.Name == name {
			out = f
			isSet = true
		}
	})
	if out == nil {
		out = fs.Lookup(name)
	}
	return
}

func (p *Loader) processEnv(config any, prefix string) error {
	configVal := reflect.Indirect(reflect.ValueOf(config))
	configTyp := configVal.Type()

	switch configTyp.Kind() {
	case reflect.Struct:
		for i := 0; i < configTyp.NumField(); i++ {
			field := configTyp.Field(i)
			fieldVal := configVal.Field(i)
			if !fieldVal.CanAddr() || !fieldVal.CanInterface() {
				continue
			}
			if field.Tag.Get(ConfrTag) == "-" {
				continue
			}

			var envNames []string
			envTag := field.Tag.Get(EnvTag)
			if envTag != "" {
				for _, name := range strings.Split(envTag, ",") {
					name = strings.TrimSpace(name)
					if name != "" {
						envNames = append(envNames, name)
					}
				}
			} else if p.EnableImplicitEnv {
				tmp := p.getEnvName(prefix, field.Name)
				envNames = append(envNames, tmp, strings.ToUpper(tmp))
			}
			if len(envNames) > 0 {
				if p.Verbose {
					p.getLogFunc()("loading env for field %s.%s from %v", configTyp.Name(), field.Name, envNames)
				}

				for _, envName := range envNames {
					if value := os.Getenv(envName); value != "" {
						err := assignFieldValue(fieldVal, value)
						if err != nil {
							return fmt.Errorf("cannot assign env value to field %s.%s: %w", configTyp.Name(), field.Name, err)
						}
						break
					}
				}
			}

			fieldVal = reflect.Indirect(fieldVal)
			switch fieldVal.Kind() {
			case reflect.Struct:
				fieldPrefix := p.getEnvName(prefix, field.Name)
				if err := p.processEnv(fieldVal.Addr().Interface(), fieldPrefix); err != nil {
					return err
				}
			case reflect.Slice:
				// Recursively process struct elements in the slice
				for i := 0; i < fieldVal.Len(); i++ {
					elemVal := reflect.Indirect(fieldVal.Index(i))
					if elemVal.Kind() == reflect.Struct {
						if err := p.processEnv(elemVal.Addr().Interface(), prefix); err != nil {
							return err
						}
					}
				}
			}
		}
	case reflect.Slice:
		// For slice types, recursively process each element
		for i := 0; i < configVal.Len(); i++ {
			elemVal := reflect.Indirect(configVal.Index(i))
			if elemVal.Kind() == reflect.Struct {
				if err := p.processEnv(elemVal.Addr().Interface(), prefix); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (p *Loader) getEnvName(prefix string, name string) string {
	var envName []byte
	for i := 0; i < len(name); i++ {
		if i > 0 && unicode.IsUpper(rune(name[i])) &&
			name[i-1] != '_' &&
			((i+1 < len(name) && unicode.IsLower(rune(name[i+1]))) || unicode.IsLower(rune(name[i-1]))) {
			envName = append(envName, '_')
		}
		envName = append(envName, name[i])
	}
	if prefix == "" {
		prefix = p.EnvPrefix
		if prefix == "" {
			prefix = DefaultEnvPrefix
		}
	}
	return prefix + "_" + string(envName)
}

func (p *Loader) processCustom(config any) error {
	if p.CustomLoader == nil {
		return nil
	}
	configVal := reflect.Indirect(reflect.ValueOf(config))
	configTyp := configVal.Type()

	switch configTyp.Kind() {
	case reflect.Struct:
		for i := 0; i < configTyp.NumField(); i++ {
			field := configTyp.Field(i)
			fieldVal := configVal.Field(i)
			if !fieldVal.CanAddr() || !fieldVal.CanInterface() {
				continue
			}
			if field.Tag.Get(ConfrTag) == "-" {
				continue
			}

			customTag := field.Tag.Get(CustomTag)
			if customTag != "" && customTag != "-" {
				if p.Verbose {
					p.getLogFunc()("processing custom loader for field %s.%s", configTyp.Name(), field.Name)
				}

				tmp, err := p.CustomLoader(fieldVal.Type(), customTag)
				if err != nil {
					return err
				}
				if err = assignFieldValue(fieldVal, tmp); err != nil {
					return fmt.Errorf("cannot assign custom value to field %s.%s: %w", configTyp.Name(), field.Name, err)
				}
			}

			fieldVal = reflect.Indirect(fieldVal)
			switch fieldVal.Kind() {
			case reflect.Struct:
				if err := p.processCustom(fieldVal.Addr().Interface()); err != nil {
					return err
				}
			case reflect.Slice:
				// Recursively process struct elements in the slice
				for i := 0; i < fieldVal.Len(); i++ {
					elemVal := reflect.Indirect(fieldVal.Index(i))
					if elemVal.Kind() == reflect.Struct {
						if err := p.processCustom(elemVal.Addr().Interface()); err != nil {
							return err
						}
					}
				}
			}
		}
	case reflect.Slice:
		// For slice types, recursively process each element
		for i := 0; i < configVal.Len(); i++ {
			elemVal := reflect.Indirect(configVal.Index(i))
			if elemVal.Kind() == reflect.Struct {
				if err := p.processCustom(elemVal.Addr().Interface()); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func assignFlagValue(dst reflect.Value, ff *flag.Flag, isSet bool) error {
	if isSet {
		if getter, ok := ff.Value.(flag.Getter); ok {
			return assignFieldValue(dst, getter.Get())
		}
		return assignFieldValue(dst, ff.Value.String())
	}

	// default value
	if dst.IsZero() && ff.DefValue != "" {
		return assignFieldValue(dst, ff.DefValue)
	}
	return nil
}

func assignFieldValue(dst reflect.Value, value any) error {
	inputVal := reflect.ValueOf(value)
	if dst.Type() == inputVal.Type() {
		dst.Set(inputVal)
		return nil
	}

	var ptrDest reflect.Value
	if dst.Kind() == reflect.Ptr {
		ptrDest = dst
		dst = reflect.New(dst.Type().Elem()).Elem()
		if dst.Type() == inputVal.Type() {
			dst.Set(inputVal)
			ptrDest.Set(dst.Addr())
			return nil
		}
	}

	var err error
	var val any
	switch dst.Interface().(type) {
	case bool:
		val, err = toBooleanE(value)
	case int:
		val, err = cast.ToIntE(value)
	case []int:
		val, err = cast.ToIntSliceE(value)
	case int64:
		val, err = cast.ToInt64E(value)
	case []int64:
		val, err = toIntSlice[int64](value)
	case int32:
		val, err = cast.ToInt32E(value)
	case []int32:
		val, err = toIntSlice[int32](value)
	case float64:
		val, err = cast.ToFloat64E(value)
	case float32:
		val, err = cast.ToFloat32E(value)
	case string:
		val, err = cast.ToStringE(value)
	case []string:
		val, err = cast.ToStringSliceE(value)
	case map[string]bool:
		val, err = cast.ToStringMapBoolE(value)
	case map[string]int:
		val, err = cast.ToStringMapIntE(value)
	case map[string]int64:
		val, err = cast.ToStringMapInt64E(value)
	case map[string]string:
		val, err = cast.ToStringMapStringE(value)
	case map[string][]string:
		val, err = cast.ToStringMapStringSliceE(value)
	case map[string]any:
		val, err = cast.ToStringMapE(value)
	case time.Duration:
		val, err = cast.ToDurationE(value)
	case []time.Duration:
		val, err = cast.ToDurationSliceE(value)
	default:
		err = errors.New("unsupported type")
	}
	if err != nil {
		return err
	}

	dst.Set(reflect.ValueOf(val))
	if ptrDest.IsValid() {
		ptrDest.Set(dst.Addr())
	}
	return nil
}

func toBooleanE(v any) (bool, error) {
	if strval, ok := v.(string); ok {
		switch strval {
		case "", "0", "f", "false", "no", "off":
			return false, nil
		case "1", "t", "true", "yes", "on":
			return true, nil
		}
	}
	return cast.ToBoolE(v)
}

func toIntSlice[T ~int32 | ~int64](v any) ([]T, error) {
	intValues, err := cast.ToIntSliceE(v)
	if err != nil {
		return nil, err
	}
	out := make([]T, len(intValues))
	for i, x := range intValues {
		out[i] = T(x)
	}
	return out, nil
}
