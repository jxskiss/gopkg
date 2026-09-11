# ezkv

Typed model caching over a caller-provided key-value `Storage`, with batch
operations, an optional Loader for cache misses, and an optional local LRU cache.

## Construction and contracts

`NewModelCache` returns `(*ModelCache[K, V], error)`. It rejects a nil config,
missing `Storage`, `IDFunc`, or `KeyFunc`, and interface model types such as
`V = ezkv.Model`. Use a concrete model type whose `UnmarshalBinary` works on a
newly allocated zero-valued element (for pointers) or on its zero value.
The constructor validates without invoking callbacks; the `Storage` callback
must return a non-nil implementation at runtime.

The constructor fills defaults in the supplied config and retains its pointer.
Do not modify the config while the cache is in use. Nonpositive batch sizes use
100 for storage operations and 300 for Loader calls. Nonpositive `LRUExpiration`
uses five seconds. `CacheExpiration` defaults to zero, meaning no expiration in
Storage. The LRU TTL is independent and can outlive the stored value.

- Returned models are **read-only**, including nested pointers, maps, and slices.
  Callers must make a deep copy before modifying them. Models supplied to write
  methods or returned by Loader must remain immutable after being handed to the
  cache; LRU entries and asynchronous writeback may retain them.
- `IDFunc` must preserve the primary key across serialization. Explicit keys in
  `Set`, `BatchSetMap`, and Loader results must match `IDFunc(model)`. Models must
  not be nil. `KeyFunc` must produce stable, distinct keys for distinct IDs.
- `Storage.Get` returns one value per requested key in the same order, with nil
  or empty values for misses, and no values on error. Returned buffers must stay
  valid and unmodified. Storage writes must consume or copy input buffers and
  slices before returning without modifying them. Missing keys are valid deletes.
- Loader results contain only requested keys and omit missing models. Loader must
  not mutate the returned map or models. Concurrent misses are not coalesced.
- Batch reads deduplicate keys and omit missing values. `BatchGetSlice` does not
  guarantee result order. Without a Loader, `Get` returns decoding errors;
  batch reads log and skip invalid values. With a Loader, read and decoding errors
  trigger fallback. Loader errors are returned; synchronous writeback errors
  accompany valid data and can be identified with `IsSetCacheError`.
- Batch writes and deletes update the LRU after each successful storage batch.
  Errors do not roll back successful batches. A failing batch can partially
  change Storage while leaving its LRU entries unchanged. Concurrent operations
  do not provide transactional consistency between the two cache layers.
- Asynchronous Loader writeback uses the request context, so cancellation may
  prevent it. Writeback errors go to `ErrorLogger`. Shared caches require their
  configured callbacks and components to support concurrent use.

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
cache, err := ezkv.NewModelCache(&ezkv.ModelCacheConfig[int64, *MyModel]{
    Storage:    storageFunc,
    IDFunc:     func(m *MyModel) int64 { return m.ID },
    KeyFunc:    keyFunc,
    Compressor: c,
})
if err != nil {
    return err
}
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
