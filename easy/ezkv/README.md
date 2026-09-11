# ezkv

Typed model caching over a caller-provided key-value `Storage`, with batch
operations, an optional Loader for cache misses, and an optional local LRU cache.

## ModelCache compression

Models implement `MarshalBinary` and `UnmarshalBinary`. Configure
`ModelCacheConfig.Compressor` to compress serialized values on writes and decode
them on reads. The default [compress](../../utils/compress/README.md)
configuration supports mixed raw values and compressed frames.

The following example assumes `MyModel` implements `ezkv.Model`, and that
`storageFunc` and `keyFunc` provide your storage and cache-key mapping:

```go
c, err := compress.NewCompressor(compress.DefaultConfig())
if err != nil {
    return err
}
cache := ezkv.NewModelCache(&ezkv.ModelCacheConfig[int64, *MyModel]{
    Storage:    storageFunc,
    IDFunc:     func(m *MyModel) int64 { return m.ID },
    KeyFunc:    keyFunc,
    Compressor: c,
})
```

The write policy applies to `Set`, `BatchSetSlice`, `BatchSetMap`, and Loader
writeback. Encoder failures use the compressor's uncompressed fallback instead
of failing the cache write; configure `CompressCallback` to observe them.

### Disabling compression writes

Set `DisableCompressionWrites: true` when constructing the cache to store
`MarshalBinary` output directly while continuing to decode existing compressed
values. The default is false.

| Configuration | Writes to Storage | Reads from Storage |
| --- | --- | --- |
| `Compressor: nil` | Raw serialized data | No decompression |
| Compressor configured | `Compress(...).Data` | Decode with the configured compressor |
| Compressor configured, `DisableCompressionWrites: true` | Raw serialized data | Decode with the configured compressor |

Keep the compressor and any required decoders configured while stored frames
still need to be read. Raw writes require a compressor accepting unframed data;
do not combine them with `compress.CompressorConfig.DisableUnframed = true`.
Create a new cache instance to change write policy rather than mutating
configuration during concurrent use.

Bypassed writes do not invoke the compressor or its callback, add framing, or
escape frame identifiers. A raw value starting with the compressor's frame
identifier can therefore be misinterpreted on reads. See the
[format detection limitations](../../utils/compress/README.md#frame-format).

### Empty values and buffer ownership

ModelCache treats zero-length values returned by Storage as cache misses.
With the standard compressor, empty serialized values are stored as nonempty
uncompressed frames and decoded before `UnmarshalBinary`. With compression
writes disabled or no compressor configured, empty serialized values remain
cache misses.

ModelCache adds no copies around serialization or storage. Respect their buffer
ownership contracts, including the compressor's
[shared-memory behavior](../../utils/compress/README.md#results-and-safe-use).
