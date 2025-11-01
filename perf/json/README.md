# json

Package json provides an on-the-fly change-able API for JSON serialization.

## Compatibility

By default, it uses [jsoniter] `ConfigCompatibleWithStandardLibrary` in underlying,
but the underlying implementation can be changed on-the-fly, e.g.
use the standard library or use a custom jsoniter config,
or switch to a [bytedance/sonic] implementation.
You may see [_examples/perf/json/bytedance_sonic]()
for an example to use bytedance/sonic as the underlying implementation.

When encoding data using `interface{}` as map keys (e.g. `map[any]any`),
both the standard library and sonic will fail, user should use jsoniter.

[bytedance/sonic]: https://github.com/bytedance/sonic

[jsoniter]: https://github.com/json-iterator/go

## Performance

By default, this package uses `jsoniter.ConfigCompatibleWithStandardLibrary` API.
It gives better performance than `encoding/json` and good compatibility with it.

User may use `ChangeImpl` to switch to a different underlying implementation.

For best performance, user may use `MarshalFastest` when the underlying
implementation is jsoniter or sonic. The result is not compatible with std
`encoding/json` in some ways, especially that map keys are not sorted.

### Benchmark

See https://github.com/json-iterator/go#benchmark.

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
3. `Fdump(w io.Writer, v any, prefix, indent string) error`

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
   benchmarking cases.

3. https://github.com/goccy/go-json <br>
   Fast JSON encoder/decoder announced to be fully compatible with encoding/json for Go.
