package bbp

import (
	"io"
	"unicode/utf8"

	"github.com/jxskiss/gopkg/v2/internal/unsafeheader"
)

// MinRead is the minimum slice size passed to a Read call by Buffer.ReadFrom.
const MinRead = 512

// NewBuffer creates a new Buffer with specified capacity.
// When you finish the work with the buffer, you may call PutBuffer
// to put it back to the pool for reusing.
func NewBuffer(capacity int) *Buffer {
	return &Buffer{buf: Get(0, capacity)}
}

// PutBuffer puts back a Buffer to the pool for reusing.
//
// The buffer mustn't be touched after returning it to the pool,
// otherwise data races will occur.
func PutBuffer(buf *Buffer) {
	put(buf.buf)
}

// Buffer provides byte buffer, which can be used for minimizing
// memory allocations.
//
// Buffer may be used with functions appending data to the underlying
// []byte slice. See example code for details.
//
// Use NewBuffer for obtaining a buffer with specified capacity.
// The zero value for Buffer is an empty buffer ready to use.
type Buffer struct {
	buf []byte
}

// Len returns the size of the byte buffer.
func (b *Buffer) Len() int {
	return len(b.buf)
}

// Grow grows the buffer's capacity, if necessary, to guarantee space for
// another n bytes. After Grow(n), at least n bytes can be written to the
// buffer without another allocation.
// If n is negative, Grow will panic.
func (b *Buffer) Grow(n int) {
	if n < 0 {
		panic("bbp.Buffer.Grow: negative size to grow")
	}
	if newCap := len(b.buf) + n; newCap > cap(b.buf) {
		b.buf = grow(b.buf, newCap, true)
	}
}

// Append accepts a function which append data to the underlying byte slice.
func (b *Buffer) Append(f func([]byte) []byte) {
	b.buf = f(b.buf)
}

// ReadFrom implements io.ReaderFrom.
//
// The function appends all the data read from r to b.
func (b *Buffer) ReadFrom(r io.Reader) (int64, error) {
	p := b.buf
	nStart := len(p)
	nMax := cap(p)
	n := nStart
	if nMax == 0 || nMax == minBufSize {
		nMax = MinRead
		p = get(nMax, nMax)
	} else {
		p = p[:nMax]
	}
	for {
		if n == nMax {
			nMax *= 2
			p = grow(p, nMax, true)
			p = p[:nMax]
		}
		nn, err := r.Read(p[n:])
		n += nn
		if err != nil {
			b.buf = p[:n]
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
	n, err := w.Write(b.buf)
	return int64(n), err
}

// Write implements io.Writer - it appends p to the underlying byte buffer.
func (b *Buffer) Write(p []byte) (int, error) {
	b.buf = append(b.buf, p...)
	return len(p), nil
}

// WriteByte appends the byte c to the buffer.
// The purpose of this function is bytes.Buffer compatibility.
func (b *Buffer) WriteByte(c byte) error {
	b.buf = append(b.buf, c)
	return nil
}

// WriteRune appends the UTF-8 encoding of Unicode code point r to the buffer.
// The purpose of this function is bytes.Buffer compatibility.
func (b *Buffer) WriteRune(r rune) (n int, err error) {
	oldLen := len(b.buf)
	b.buf = utf8.AppendRune(b.buf, r)
	return len(b.buf) - oldLen, nil
}

// WriteString appends s to the underlying byte slice.
func (b *Buffer) WriteString(s string) (int, error) {
	b.buf = append(b.buf, s...)
	return len(s), nil
}

// WriteStrings appends a slice of strings to the underlying byte slice.
func (b *Buffer) WriteStrings(s []string) (int, error) {
	lenb := len(b.buf)
	lens := 0
	for i := 0; i < len(s); i++ {
		lens += len(s[i])
	}
	want := lenb + lens
	if want > cap(b.buf) {
		b.buf = grow(b.buf, want, true)
	}
	b.buf = b.buf[:want]
	for _, x := range s {
		lenb += copy(b.buf[lenb:], x)
	}
	return lens, nil
}

// Set first re-slice the underlying byte slice to empty,
// then write p to the buffer.
func (b *Buffer) Set(p []byte) {
	b.buf = append(b.buf[:0], p...)
}

// SetString first re-slice the underlying byte slice to empty,
// then write s to the buffer.
func (b *Buffer) SetString(s string) {
	b.buf = append(b.buf[:0], s...)
}

// Reset re-slice the underlying byte slice to empty.
func (b *Buffer) Reset() {
	b.buf = b.buf[:0]
}

// Clone returns a new copy of the buffer, including the underlying
// byte slice.
func (b *Buffer) Clone() *Buffer {
	cp := &Buffer{
		buf: make([]byte, len(b.buf), cap(b.buf)),
	}
	copy(cp.buf, b.buf)
	return cp
}

// Bytes returns the underlying byte slice, i.e. all the bytes accumulated
// in the buffer.
//
// Note that this method doesn't copy the underlying byte slice, the caller
// should either copy the byte slice explicitly or don't return the Buffer
// back to the pool, otherwise data race will occur.
// You may use Buffer.Copy to get a copy of the underlying byte slice.
func (b *Buffer) Bytes() []byte {
	return b.buf
}

// Copy returns a copy of the underlying byte slice.
func (b *Buffer) Copy() []byte {
	buf := make([]byte, len(b.buf))
	copy(buf, b.buf)
	return buf
}

// String returns a string copy of the underlying byte slice.
func (b *Buffer) String() string {
	return string(b.buf)
}

// StringUnsafe is equivalent to String, but the string that it returns
// is _NOT_ copied, so modifying this buffer after calling StringUnsafe
// will lead to undefined behavior.
func (b *Buffer) StringUnsafe() string {
	return unsafeheader.BytesToString(b.buf)
}
