Package bbp provides efficient byte buffer pools with anti-memory-waste protection.

Byte buffers acquired from this package may be put back to the pool, but they do not need to;
if they are returned, they will be recycled and reused, otherwise they will be garbage
collected as usual.
