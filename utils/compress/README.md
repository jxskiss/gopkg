# compress

Optional compression for in-memory values such as cache entries. Small or
poorly compressible values are stored uncompressed; decoding accepts both
compressed frames and raw (unframed) data by default.

## Usage

```go
cfg := compress.DefaultConfig()
cfg.Threshold = 1024
cfg.MaxDecodedSize = 8 << 20
c, err := compress.NewCompressor(cfg)
if err != nil {
    return err
}

result := c.Compress(ctx, original)
stored := result.Data
decoded, algorithm, err := c.Decompress(ctx, stored)
if err != nil {
    return err
}
```

Always store `result.Data` and pass it back to `Decompress`; callers do not need
to distinguish compressed and uncompressed representations. Even if compression
fails, `Data` remains usable as an uncompressed fallback. Check `result.Err` if
your application must reject that fallback.

## Configuration

Start with `DefaultConfig()` and adjust it for your workload.

| Option | Default | Purpose |
| --- | --- | --- |
| `Codec` | gzip, standard default level | Algorithm used for compression |
| `Threshold` | 5 KiB | Minimum input size eligible for compression |
| `MinReductionRatio` | 5% | Required size saving, including the frame header |
| `MaxDecodedSize` | 64 MiB | Maximum output size for actual decompression |
| `DisableUnframed` | `false` | Always write frames and reject raw input when decoding |
| `Decoders` | None beyond the selected codec | Additional algorithms accepted when reading |
| `CompressCallback` | `nil` | Observe compression results, including skips and failures |

Zero `Threshold` removes the size threshold. Zero `MinReductionRatio` still
requires the complete compressed frame to be smaller than the original input.
Zero `MaxDecodedSize` selects the default limit. Choose thresholds and codec
settings using representative data; the defaults are starting points.

`MaxDecodedSize` applies only to actual decompression, not to raw data or
uncompressed (`TypeNone`) frames. It does not limit input to `Compress`, so
reading a compressed value larger than the limit requires a larger decoder limit.

## Results and safe use

- `CompressionResult.Reason` explains why compression was skipped or failed.
  Only encoder failure sets `Err`; the package does not log automatically.
  Use `CompressCallback` for metrics or logging.
- Methods do not modify input, but results may share its underlying array.
  Keep shared buffers unchanged and out of pools until the result is consumed,
  or use `bytes.Clone` for independent storage. Later compressor calls do not
  invalidate returned data.
- Empty input is stored as a nonempty uncompressed frame. Nil and empty slices
  are not distinguished after this round trip. Uncompressed input that matches
  the frame identifier is also wrapped to preserve round trips.
- Decode errors return nil data. Automatic format detection is not an integrity
  check; see [Frame format](#frame-format) for its limitations.
- Concurrent use requires concurrency-safe codecs and callbacks. Context is
  passed to the callback only; operations do not support cancellation.

## Choosing codecs

Gzip is built in. Set `Codec` to `NewGzipCodec(level)` to choose a compression
level; standard gzip levels are supported, including `gzip.NoCompression` (0).

The selected codec also handles decoding its own frames. When changing write
algorithms, add decoders for existing data to `CompressorConfig.Decoders`.
For example, a gzip writer needs a zstd decoder to read existing zstd frames;
a zstd writer needs a gzip decoder to read existing gzip frames.

The [zstd example](../../_examples/utils/compress-impl/zstd) demonstrates a
custom `Codec` without adding a zstd dependency to this package. See the
[codec interfaces](alg.go) for implementation contracts. To decode bare gzip
or zstd streams, use the codec API directly rather than `Compressor`.

## Frame format

The format identifier is a fixed 128-bit magic, stored as the following binary
bytes (not the hexadecimal text):

```text
13 39 ce 30 c0 5b 98 c4 4e 13 21 15 35 55 c4 b8
```

It is the same constant across instances and processes. Version 1 has a 26-byte
header with the layout below.

| Offset | Size | Meaning |
| --- | --- | --- |
| 0 | 16 | Fixed magic |
| 16 | 1 | Version, currently 1 |
| 17 | 1 | Algorithm: ASCII `0` (uncompressed), `1` (gzip), or `2` (zstd) |
| 18 | 8 | Uncompressed length, unsigned big-endian |
| 26 | remaining | Uncompressed or compressed payload |

Empty values also have a frame. Decoding validates the header, version, algorithm,
declared length and actual output. The declared length is never trusted for an
unbounded allocation; codecs enforce the configured limit during decompression.
The limit covers decompressed bytes, not uncompressed input, total process memory
or codec working space.
The frame provides length checks, not authentication or a checksum for uncompressed
payloads. Underlying codec checksum guarantees still apply.

Only the complete 16-byte magic selects frame decoding. Once it matches,
truncation, unknown versions or algorithms, length mismatches and decompression
failures return errors; they never fall back to unframed data. A `TypeNone` frame
is unwrapped once, without interpreting magic inside its payload again.

Without a full magic match, the input is returned unchanged by default, even if
it contains a partial magic prefix or a bare gzip/zstd stream. Use the explicit
codec API to decode bare streams. `DisableUnframed` rejects all such input.

The long magic reduces accidental collisions with externally written unframed
data; it does not make them impossible. `Compress` escapes matching uncompressed
input in a `TypeNone` frame. Damage to the magic itself can make a frame appear
unframed, so automatic detection is not an integrity check.

## Verification and tuning

```sh
go test -race -ldflags="-linkmode=external" ./utils/compress
go test -ldflags="-linkmode=external" ./utils/compress -run '^$' -bench . -benchmem
go test -ldflags="-linkmode=external" ./utils/compress -run '^$' -fuzz '^FuzzCompressorDecompress$' -fuzztime=30s
go test -ldflags="-linkmode=external" ./utils/compress -run '^$' -fuzz '^FuzzCompressorRoundTrip$' -fuzztime=30s
```

Benchmarks cover short text, JSON, random data, already-compressed data, decoding,
and concurrent round trips, including unframed and `TypeNone` decoding.
`frame-ratio` compares actual output size, including a header when present,
against the original input size. Defaults are starting points, not a
claim of optimal performance; choose codec, level and thresholds using real data.
