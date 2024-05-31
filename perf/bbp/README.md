Package bbp provides efficient byte buffer pools with anti-memory-waste protection.

Byte buffers acquired from this package may be put back to the pool, but they do not need to;
if they are returned, they will be recycled and reused, otherwise they will be garbage
collected as usual.

The methods within this package and all `Pool` instances share the same
underlying sized byte slice pools. The byte buffers provided by this package
has a minimum limit of 64B and a maximum limit of 4MB,
byte slice with size not in the range will be allocated directly from Go runtime,
and won't be recycled for reuse.
