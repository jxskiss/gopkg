package bbp

import (
	"io"
	"sync"
	"unicode/utf8"
)

var bpool = sync.Pool{
	New: func() interface{} { return &Buffer{} },
}

// NewBuffer creates and initializes a new Buffer using buf as its
// initial contents. The new Buffer takes ownership of buf, and the
// caller should not use buf after this call. NewBuffer is intended to
// prepare a Buffer to read existing data. It can also be used to set
// the initial size of the internal buffer for writing. To do that,
// buf should have the desired capacity but a length of zero.
//
// In most cases, Get(length, capacity), new(Buffer), or just declaring
// a Buffer variable is sufficient to initialize a Buffer.
func NewBuffer(buf []byte) *Buffer {
	b := bpool.Get().(*Buffer)
	b.noReuse = true
	b.B = buf
	return b
}

// Buffer provides byte buffer, which can be used for minimizing
// memory allocations.
//
// Buffer may be used with functions appending data to the given []byte
// slice. See example code for details.
//
// Use Get for obtaining an empty byte buffer.
// The zero value for Buffer is ready to use.
type Buffer struct {

	// B is a byte buffer to use in append-like workloads.
	// See example code for details.
	B []byte

	noReuse bool
}

// Len returns the size of the byte buffer.
func (b *Buffer) Len() int {
	return len(b.B)
}

// 512 is considered a reasonable default size to work with io.Reader streams
const defaultBufSizeForIOReader = 512

// ReadFrom implements io.ReaderFrom.
//
// The function appends all the data read from r to b.
func (b *Buffer) ReadFrom(r io.Reader) (int64, error) {
	bb := b.B
	nStart := len(bb)
	nMax := cap(bb)
	n := nStart
	if nMax == 0 {
		nMax = defaultBufSizeForIOReader
		bb = get(nMax)
	} else {
		bb = bb[:nMax]
	}
	for {
		if n == nMax {
			nMax *= 2
			bb = grow(bb, nMax)
			bb = bb[:nMax]
		}
		nn, err := r.Read(bb[n:])
		n += nn
		if err != nil {
			b.B = bb[:n]
			n -= nStart
			if err == io.EOF {
				return int64(n), nil
			}
			return int64(n), err
		}
	}
}

// WriteTo implements io.WriterTo.
func (b *Buffer) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write(b.B)
	return int64(n), err
}

// Write implements io.Writer - it appends p to the underlying byte buffer.
func (b *Buffer) Write(p []byte) (int, error) {
	lenb, lenp := len(b.B), len(p)
	want := lenb + lenp
	if want > cap(b.B) {
		b.B = grow(b.B, want)[:want]
	}
	copy(b.B[lenb:], p)
	return lenp, nil
}

// WriteByte appends the byte c to the buffer.
//
// The purpose of this function is bytes.Buffer compatibility.
//
// The function always returns nil.
func (b *Buffer) WriteByte(chars ...byte) error {
	lenb, lenc := len(b.B), len(chars)
	if lenc == 0 {
		return nil
	}
	want := lenb + len(chars)
	if want > cap(b.B) {
		b.B = grow(b.B, want)[:want]
	}
	copy(b.B[lenb:], chars)
	return nil
}

// WriteRune appends the UTF-8 encoding of Unicode code point r to the buffer.
//
// The purpose of this function is bytes.Buffer compatibility.
//
// The function always returns nil.
func (b *Buffer) WriteRune(r rune) (n int, err error) {
	if r < utf8.RuneSelf {
		b.WriteByte(byte(r))
		return 1, nil
	}
	lenb := len(b.B)
	want := lenb + utf8.UTFMax
	if want > cap(b.B) {
		b.B = grow(b.B, want)
	}
	n = utf8.EncodeRune(b.B[lenb:lenb+utf8.UTFMax], r)
	b.B = b.B[:lenb+n]
	return n, nil
}

// WriteString appends s to the underlying byte slice.
func (b *Buffer) WriteString(s string) (int, error) {
	lenb, lens := len(b.B), len(s)
	want := lenb + lens
	if want > cap(b.B) {
		b.B = grow(b.B, want)[:want]
	}
	copy(b.B[lenb:], s)
	return lens, nil
}

// Set first re-slice the underlying byte slice to emtpy,
// then write p to the buffer.
func (b *Buffer) Set(p []byte) {
	b.B = b.B[:0]
	b.Write(p)
}

// SetString sets Buffer.B to s.
func (b *Buffer) SetString(s string) {
	b.B = b.B[:0]
	b.WriteString(s)
}

// Reset re-slice the underlying byte slice to empty.
func (b *Buffer) Reset() {
	b.B = b.B[:0]
}

// Bytes returns b.B, i.e. all the bytes accumulated in the buffer.
//
// Note that this method doesn't copy the underlying byte slice, the caller
// should either copy the byte slice explicitly or don't return the Buffer
// back to the pool, otherwise data race will occur.
// If you want a copy of the underlying byte slice, you can use Buffer.Copy
// or copy Buffer.B manually.
//
// The purpose of this function is bytes.Buffer compatibility.
func (b *Buffer) Bytes() []byte {
	return b.B
}

// Copy returns a copy of the underlying byte slice.
func (b *Buffer) Copy() []byte {
	buf := make([]byte, len(b.B))
	copy(buf, b.B)
	return buf
}

// String returns a string copy of the underlying byte slice.
func (b *Buffer) String() string {
	return string(b.B)
}