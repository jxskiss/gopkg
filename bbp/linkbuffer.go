package bbp

import (
	"io"
	"unicode/utf8"
)

const (
	linkBufferMinPoolIdx   = 10
	linkBufferMinBlockSize = 1 << linkBufferMinPoolIdx // 1024

	linkBufferDefaultPoolIdx   = 12
	linkBufferDefaultBlockSize = 1 << linkBufferDefaultPoolIdx // 4096
)

// NewLinkBuffer creates and initializes a new LinkBuffer, blockSize will
// be round up to the next power of two, and limited in between MinSize
// and MaxSize, if bockSize <= 0, a default value 512 will be used.
func NewLinkBuffer(blockSize int) *LinkBuffer {
	if blockSize <= 0 {
		blockSize = linkBufferDefaultBlockSize
	}
	if blockSize < linkBufferMinBlockSize {
		blockSize = linkBufferMinBlockSize
	}
	if blockSize > MaxSize {
		blockSize = MaxSize
	}

	poolIdx := indexGet(blockSize)
	blockSize = 1 << poolIdx

	buf := &LinkBuffer{
		blockSize: blockSize,
		poolIdx:   poolIdx,
	}
	return buf
}

// LinkBuffer provides a linked buffer, which can be used to minimizing
// memory allocation. It's recommended to get a LinkBuffer by calling
// NewLinkBuffer with proper block size which helps to reduce memory
// allocation.
//
// LinkBuffer may be used with functions appending data to the underlying
// []byte slices. See example code for details.
//
// The zero value for LinkBuffer is an empty buffer ready to use.
type LinkBuffer struct {
	blockSize int
	poolIdx   int

	tmp  [8]byte
	bufs [][]byte
	size int
	cap  int
}

func (b *LinkBuffer) getBlockSize() int {
	if b.blockSize > 0 {
		return b.blockSize
	}
	return linkBufferDefaultBlockSize
}

func (b *LinkBuffer) getPoolIdx() int {
	if b.poolIdx > 0 {
		return b.poolIdx
	}
	return linkBufferDefaultPoolIdx
}

func (b *LinkBuffer) grow(incr int) {
	blockSize := b.getBlockSize()
	poolIdx := b.getPoolIdx()
	n := (incr + b.size - b.cap + blockSize - 1) / blockSize
	for i := 0; i < n; i++ {
		buf := sizedPools[poolIdx].Get().([]byte)
		b.bufs = append(b.bufs, buf)
		b.cap += blockSize
	}
}

// Len returns the size of the LinkBuffer.
func (b *LinkBuffer) Len() int {
	return b.size
}

// ReadFrom implements io.ReaderFrom.
//
// The function appends all the data read from r to b.
func (b *LinkBuffer) ReadFrom(r io.Reader) (int64, error) {
	blockSize := b.getBlockSize()
	want := b.size + blockSize
	if want > b.cap {
		b.grow(blockSize)
	}
	n := 0
	idx := b.size / blockSize
	for {
		bb := b.bufs[idx]
		i := len(bb)
		nn, err := r.Read(bb[i:cap(bb)])
		n += nn
		b.size += nn
		b.bufs[idx] = bb[:i+nn]
		if err != nil {
			if err == io.EOF {
				return int64(n), nil
			}
			return int64(n), err
		}
		b.grow(blockSize)
		idx += 1
	}
}

// WriteTo implements io.WriterTo.
func (b *LinkBuffer) WriteTo(w io.Writer) (int64, error) {
	var n int
	for _, bb := range b.bufs {
		if len(bb) == 0 {
			break
		}
		nn, err := w.Write(bb)
		n += nn
		if err != nil {
			return int64(n), err
		}
	}
	return int64(n), nil
}

// Write implements io.Writer - it appends p to the buffer.
func (b *LinkBuffer) Write(p []byte) (int, error) {
	return b.WriteString(b2s(p))
}

// WriteByte appends the byte c to the buffer.
//
// The purpose of this function is bytes.Buffer compatibility.
//
// The function always returns nil.
func (b *LinkBuffer) WriteByte(c byte) error {
	blockSize := b.getBlockSize()
	if b.size == b.cap {
		b.grow(1)
	}
	idx := b.size / blockSize
	b.bufs[idx] = append(b.bufs[idx], c)
	b.size += 1
	return nil
}

// WriteRune appends the UTF-8 encoding of Unicode code point r to the buffer.
//
// The purpose of this function is bytes.Buffer compatibility.
//
// The function always returns nil.
func (b *LinkBuffer) WriteRune(r rune) (n int, err error) {
	tmp := b.tmp[:4]
	n = utf8.EncodeRune(tmp, r)
	b.WriteString(b2s(tmp[:n]))
	return
}

// WriteString appends s to the LinkBuffer.
func (b *LinkBuffer) WriteString(s string) (int, error) {
	lens := len(s)
	if lens == 0 {
		return 0, nil
	}

	want := b.size + lens
	if want > b.cap {
		b.grow(lens)
	}

	blockSize := b.getBlockSize()
	return b.copyString(blockSize, s), nil
}

// WriteStrings appends a slice of strings to the underlying byte buffers.
func (b *LinkBuffer) WriteStrings(strs []string) (int, error) {
	lens := 0
	for i := 0; i < len(strs); i++ {
		lens += len(strs[i])
	}
	if lens == 0 {
		return 0, nil
	}

	want := b.size + lens
	if want > b.cap {
		b.grow(lens)
	}

	blockSize := b.getBlockSize()
	n := 0
	for _, s := range strs {
		n += b.copyString(blockSize, s)
	}
	return n, nil
}

func (b *LinkBuffer) copyString(blockSize int, s string) int {
	lens := len(s)
	idx := b.size / blockSize
	n := 0
	for lens > 0 {
		bb := b.bufs[idx]
		i := len(bb)
		nn := copy(bb[i:blockSize], s[n:])
		n += nn
		b.size += nn
		b.bufs[idx] = bb[:i+nn]
		idx += 1
		lens -= nn
	}
	return n
}

// Reset resets the LinkBuffer to empty and returns the underlying
// byte buffers to the pool for reusing.
func (b *LinkBuffer) Reset() {
	poolIdx := b.getPoolIdx()
	for _, bb := range b.bufs {
		sizedPools[poolIdx].Put(bb[:0])
	}
	b.bufs = nil
	b.size, b.cap = 0, 0
}

// Copy returns a copy of the underlying byte slice.
func (b *LinkBuffer) Copy() []byte {
	buf := make([]byte, b.size)
	n := 0
	for _, bb := range b.bufs {
		nn := len(bb)
		copy(buf[n:n+nn], bb)
		n += nn
	}
	return buf
}

// String returns a string copy of the underlying byte slice.
func (b *LinkBuffer) String() string {
	buf := make([]byte, b.size)
	n := 0
	for _, bb := range b.bufs {
		nn := len(bb)
		copy(buf[n:n+nn], bb)
		n += nn
	}
	return b2s(buf)
}

// Reader returns a Reader reading from the Buffer's underlying byte buffers.
//
// When the returned Reader is being used, modifying this LinkBuffer will
// lead to undefined behavior.
func (b *LinkBuffer) Reader() *Reader {
	return NewReader(b.bufs...)
}

// PutLinkBuffer puts back a LinkBuffer to the pool for reusing.
//
// The buf mustn't be touched after returning it to the pool.
// Otherwise, data races will occur.
func PutLinkBuffer(buf *LinkBuffer) {
	poolIdx := buf.poolIdx
	for _, bb := range buf.bufs {
		sizedPools[poolIdx].Put(bb[:0])
	}
	buf.bufs = nil
}
