# compress

This package adds a storage frame and a compression policy around raw codecs.
It is intended for in-memory values, such as cache entries. Context is passed to
the compression observer only; operations do not support cancellation.

## Usage

```go
cfg := compress.DefaultConfig()
cfg.Threshold = 1024
cfg.MinReductionRatio = 0.05
cfg.MaxDecodedSize = 8 << 20
c, err := compress.NewCompressor(cfg)
if err != nil {
    return err
}

result := c.Compress(ctx, original)
// result.Data is always a complete frame, including on encoder failure.
// Check result.Err here if this application must reject fallback.
frame := result.Data
decoded, algorithm, err := c.Decompress(ctx, frame)
```

`DefaultConfig` supplies a 5 KiB threshold, 5% minimum reduction, and a 64 MiB
decoded-size limit. An explicit zero threshold or reduction is respected. A zero
decoded-size limit selects the default; negative limits are invalid. Reduction
must be finite and within [0, 1). Compression is selected only when its complete
frame is strictly smaller than the original data and meets the minimum reduction
ratio: `(originalSize - compressedFrameSize) / originalSize`. For example,
`MinReductionRatio = 0.05` requires the complete compressed frame, including its
header, to be at least 5% smaller than the original data. Zero still rejects
equal-sized or expanded results. Uncompressed fallback adds a frame header, so
its stored size is larger than the original data.

`CompressionResult.Reason` distinguishes empty input, small input, expansion,
insufficient reduction, and encoder failure. Only encoder failure sets `Err`, which
wraps the original error. The fallback still contains the original data in a
valid frame. The package does not log automatically. `CompressCallback` receives
`CompressionStats` value metadata and may run concurrently; it cannot mutate
internal configuration. `CompressionStats.Algorithm` identifies the configured
codec, including skipped and failed attempts. An empty `Reason` means the result
uses that algorithm; otherwise the frame contains uncompressed data (`TypeNone`).

Compression does not reject oversized input: the decoder limit is a read policy.
A value encoded above that limit requires a decoder configured with a larger
limit. All successful high-level outputs are independent of their inputs and of
subsequent calls. Nil and empty input both encode an empty value; their distinction
is not preserved. Decode errors return nil data.

## Codecs and decoding

`NewGzipCodec(level)` preserves standard gzip level values, including
`gzip.NoCompression` (0). It returns a configuration error for invalid levels.
Codec methods operate on raw streams, without this package's storage frame.
`Codec.Compress` appends to its destination and requires non-overlapping source
and destination. `Decoder.Decompress` requires a positive decoded-size limit and
validates the complete stream, including any concatenated members.

The selected codec supplies its own decoder. Additional decoders are explicitly
passed in `CompressorConfig.Decoders`; no package import registers global state.
Configuration is copied at construction, including the decoder lookup. Codec
instances themselves must remain safe for concurrent use and keep their type
stable. Duplicate or unsupported algorithm identifiers are rejected.

The zstd implementation in `_examples/utils/compress-impl/zstd` demonstrates
implementing `Codec` without adding a zstd dependency to the core module. To read
gzip and zstd frames with a gzip encoder, supply the zstd codec in `Decoders`.
To write zstd, put it in `Codec` and explicitly add a gzip decoder if needed.

## Frame format

The format identifier is fixed as ASCII `GPKC` (gopkg compression), bytes
`47 50 4b 43` in hexadecimal. Version 1 has a 14-byte header with the layout below.

| Offset | Size | Meaning |
| --- | --- | --- |
| 0 | 4 | ASCII `GPKC` magic |
| 4 | 1 | Version, currently 1 |
| 5 | 1 | Algorithm: ASCII `0` (uncompressed), `1` (gzip), or `2` (zstd) |
| 6 | 8 | Uncompressed length, unsigned big-endian |
| 14 | remaining | Uncompressed or compressed payload |

Empty values also have a frame. Decoding validates the header, version, algorithm,
declared length and actual output. The declared length is never trusted for an
unbounded allocation; codecs also enforce the configured limit during decoding.
The limit covers output bytes, not total process memory or codec working space.
The frame provides length checks, not authentication or a checksum for uncompressed
payloads. Underlying codec checksum guarantees still apply.

Bare gzip/zstd streams and arbitrary plaintext are not accepted.
Decode bare streams with the explicit codec API. Applications migrating plaintext
storage should select a format using external version metadata instead of
guessing from payload prefixes.

## Verification and tuning

```sh
go test -race -ldflags="-linkmode=external" ./utils/compress ./easy/ezkv
go test -ldflags="-linkmode=external" ./utils/compress -run '^$' -bench . -benchmem
go test -ldflags="-linkmode=external" ./utils/compress -run '^$' -fuzz '^FuzzCompressorDecompress$' -fuzztime=30s
```

Benchmarks cover short text, JSON, random data, already-compressed data, decoding,
and concurrent round trips. `frame-ratio` compares output frame sizes, including
the header, against the original input size. Defaults are starting points, not a
claim of optimal performance; choose codec, level and thresholds using real data.
