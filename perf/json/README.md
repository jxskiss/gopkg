# json

## Compatibility

This package provides a drop-in replacement of `encoding/json`.
Any incompatible behavior with `encoding/json` is considered a bug.

To get better performance, you can build your application with tag `unsafejson`
to use [go-json], which is a fully compatible drop-in replacement
of `encoding/json` but with high performance.

[go-json]: https://github.com/goccy/go-json

## Performance

By default, this package uses `encoding/json` from the standard library.
It gives the best compatibility but lower performance.

You can build your application with tag `unsafejson` to switch to `goccy/go-json`,
which has much better performance than `encoding/json` and many other
third-party JSON libraries.

`goccy/go-json` is announced and tested as 100% drop-in replacement of `encoding/json`,
but if you encounter some incompatible behavior unfortunately, you may remove the `unsafejson`
build tag to quickly switch to `encoding/json`, and please
[submit an issue](https://github.com/goccy/go-json/issues) to `goccy/go-json`.

For marshalling map data where key ordering does not matter, you may use the shortcut
function `MarshalMapUnordered`, which disables map key ordering to get even better performance
(this is only available with `unsafejson` tag, not the standard library implementation).

## Utilities

String operation avoiding unnecessary memory allocation:

1. `MarshalToString(v interface{}) (string, error)`
2. `UnmarshalFromString(str string, v interface{}) error`

Encoder and Decoder with method chaining capabilities:

1. `NewEncoder(w).SetEscapeHTML(false).SetIndent("", "  ").Encode(v)`
2. `NewDecoder(r).UseNumber().DisallowUnknownFields().Decode(v)`

Disable HTMLEscape to get output more friendly to read for human:

1. `MarshalNoHTMLEscape(v interface{}, prefix, indent string) ([]byte, error)`

Handy shortcuts to load and dump JSON data from/to file:

1. `Load(path string, v interface{}) error`
2. `Dump(path string, v interface{}, prefix, indent string) error`

## Benchmark

See https://github.com/goccy/go-json#benchmarks.

## Other JSON libraries

1. https://github.com/tidwall/gjson <br>
   GJSON is a Go package that provides a fast and simple way to get values from a json document.
   It has features such as one line retrieval, dot notation paths, iteration, and parsing json lines.
 
2. https://github.com/bytedance/sonic <br>
   Sonic is a blazingly fast JSON serializing & deserializing library, accelerated by JIT and SIMD.
   It is not a 100% drop-in replacement of `encoding/json`, but it performs best for various
   benchmarking cases, you may use it in super hot code path (but you probably want to firstly
   review your design which use JSON in that hot path).

3. https://github.com/jxskiss/extjson <br>
   extjson extends the JSON syntax, it allows extended features such as
   trailing comma in object and array, adding comment, including another JSON file,
   referencing to other values.
