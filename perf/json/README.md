# json

## Compatibility

This package provides a wrapper implementation of `encoding/json`.

By default, it uses [jsoniter] `ConfigCompatibleWithStandardLibrary` in underlying,
but the behavior can be configured, e.g.
using the standard library or using a custom jsoniter config,
or switch to a [bytedance/sonic] config.
Note that we may change to use bytedance/sonic as default in the future,
when it's fully ready for production deployment.

When encoding data using `interface{}` as map keys (e.g. `map[any]any`),
both the standard library and sonic will fail, user should use jsoniter.

[bytedance/sonic]: https://github.com/bytedance/sonic
[jsoniter]: https://github.com/json-iterator/go

## Performance

By default, this package uses `jsoniter.ConfigCompatibleWithStandardLibrary` API.
It gives much better performance than `encoding/json` and good compatibility with it.

User may use the method `Config` to customize the behavior of this library.

For best performance, user may use `MarshalFastest` when this library is
configured to use jsoniter or sonic. The result is not compatible with
std `encoding/json` in some ways, especially that map keys are not sorted.

### Benchmark

See https://github.com/json-iterator/go#benchmark,
and https://github.com/bytedance/sonic#benchmarks.

## Utilities

String operation avoiding unnecessary memory allocation:

1. `MarshalToString(v any) (string, error)`
2. `UnmarshalFromString(str string, v any) error`

Encoder and Decoder with method chaining capabilities:

1. `NewEncoder(w).SetEscapeHTML(false).SetIndent(prefix, indent).Encode(v)`
2. `NewDecoder(r).UseNumber().DisallowUnknownFields().Decode(v)`

Disable HTMLEscape to get output more friendly to read for human:

1. `MarshalNoHTMLEscape(v any, prefix, indent string) ([]byte, error)`

Handy shortcuts to load and dump JSON data from/to a file:

1. `Load(path string, v any) error`
2. `Dump(path string, v any, prefix, indent string) error`

Generates human-friendly result (with lower performance):

1. `HumanFriendly.Marshal(v any) ([]byte, error)`
2. `HumanFriendly.MarshalToString(v any) (string, error)`
3. `HumanFriendly.MarshalIndent(v any, prefix, indent string) ([]byte, error)`
4. `HumanFriendly.MarshalIndentString(v any, prefix, indent string) (string, error)`
5. `HumanFriendly.NewEncoder(w io.Writer) *Encoder`

## Other JSON libraries

1. https://github.com/tidwall/gjson <br>
   GJSON is a Go package that provides a fast and simple way to get values from a json document.
   It has features such as one line retrieval, dot notation paths, iteration, and parsing json lines.

2. https://github.com/bytedance/sonic <br>
   Sonic is a blazingly fast JSON serializing & deserializing library, accelerated by JIT and SIMD.
   It is not a 100% drop-in replacement of `encoding/json`, but it performs best for various
   benchmarking cases, you may use it in super hot code path (but you probably want to firstly
   review your design, using JSON in hot path is considered bad practice).

3. https://github.com/goccy/go-json <br>
   Fast JSON encoder/decoder announced to be fully compatible with encoding/json for Go.

4. https://github.com/jxskiss/extjson <br>
   `extjson` extends the JSON syntax, it allows extended features such as
   trailing comma in object and array, adding comment, including another JSON file,
   referencing to other values.
