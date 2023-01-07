# json

## Compatibility

This package provides a wrapper implementation of `encoding/json`.

By default, it uses [bytedance/sonic] `ConfigStd` in underlying,
but the behavior can be configured, e.g.
using the standard library or using a custom sonic config,
or switch to a [jsoniter] config.

When you are encoding data using `interface{}` as map keys (e.g. `map[interface{}]interface{}`),
both the standard library and sonic will fail, you should use jsoniter.

[bytedance/sonic]: https://github.com/bytedance/sonic
[jsoniter]: https://github.com/json-iterator/go

## Performance

By default, this package uses sonic's `CofigStd` API.
It gives much better performance than `encoding/json` and good compatibility with it.

You may use the method `Config` to customize the behavior of this library.
In case that sonic behaves in some unexpected way, you may switch to
jsoniter or the standard library to do a quick fix.

For marshalling map data where key ordering does not matter,
you may use the shortcut function `MarshalNoMapOrdering`,
which disables map key ordering in case that sonic is used.
(Note that this is only available with sonic, not the standard library implementation.)

### Benchmark

See https://github.com/bytedance/sonic#benchmarks.

## Utilities

String operation avoiding unnecessary memory allocation:

1. `MarshalToString(v interface{}) (string, error)`
2. `UnmarshalFromString(str string, v interface{}) error`

Encoder and Decoder with method chaining capabilities:

1. `NewEncoder(w).SetEscapeHTML(false).SetIndent(prefix, indent).Encode(v)`
2. `NewDecoder(r).UseNumber().DisallowUnknownFields().Decode(v)`

Disable HTMLEscape to get output more friendly to read for human:

1. `MarshalNoHTMLEscape(v interface{}, prefix, indent string) ([]byte, error)`

Handy shortcuts to load and dump JSON data from/to a file:

1. `Load(path string, v interface{}) error`
2. `Dump(path string, v interface{}, prefix, indent string) error`

## Other JSON libraries

1. https://github.com/tidwall/gjson <br>
   GJSON is a Go package that provides a fast and simple way to get values from a json document.
   It has features such as one line retrieval, dot notation paths, iteration, and parsing json lines.

2. https://github.com/bytedance/sonic <br>
   Sonic is a blazingly fast JSON serializing & deserializing library, accelerated by JIT and SIMD.
   It is not a 100% drop-in replacement of `encoding/json`, but it performs best for various
   benchmarking cases, you may use it in super hot code path (but you probably want to firstly
   review your design which use JSON in that hot path).

3. https://github.com/goccy/go-json <br>
   Fast JSON encoder/decoder announced to be fully compatible with encoding/json for Go.

4. https://github.com/jxskiss/extjson <br>
   `extjson` extends the JSON syntax, it allows extended features such as
   trailing comma in object and array, adding comment, including another JSON file,
   referencing to other values.
