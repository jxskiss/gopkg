Package bbp provides efficient byte buffer pools with anti-memory-waste protection.

Byte buffers acquired from this package may be put back to the pool, but they do not need to;
if they are returned, they will be recycled and reused, otherwise they will be garbage
collected as usual.

Within this package, `Get`, `Set` and all `Pool` instances share the same
underlying sized byte slice pools. The byte buffers provided by this package
has minimum and maximum limit (see `MinBufSize` and `MaxBufSize`),
byte slice with size not in the range will be allocated directly from the
operating system, and won't be recycled for reuse.
