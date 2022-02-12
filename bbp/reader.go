package bbp

import (
	"errors"
	"io"
	"sort"
	"unicode/utf8"
)

// A Reader implements the io.Reader, io.ReaderAt, io.WriterTo, io.Seeker,
// io.ByteScanner, and io.RuneScanner interfaces by reading from
// a byte slice.
//
// It mimics bytes.Reader and is read-only.
type Reader struct {
	bufs     [][]byte
	offs     []int64
	size     int64 // total size of bufs
	bi       int64 // current buf index
	i        int64 // current reading index
	prevRune int64 // index of previous rune; or < 0
}

func (r *Reader) advance(n int64) {
	if r.i+n > r.size {
		panic("bbp.Reader.advance: invalid position")
	}
	r.i += n
	for int(r.bi) < len(r.bufs) && int(r.i-r.offs[r.bi]) >= len(r.bufs[r.bi]) {
		r.bi++
	}
}

func (r *Reader) moveback(n int64) {
	if r.i-n < 0 {
		panic("bbp.Reader.moveback: invalid position")
	}
	r.i -= n
	for r.i < r.offs[r.bi] {
		r.bi--
	}
}

// Len returns the number of bytes of the unread portion of the slice.
func (r *Reader) Len() int {
	if r.i >= r.size {
		return 0
	}
	return int(r.size - r.i)
}

// Size returns the original length of the underlying byte buffers.
// Size is the number of bytes available for reading via ReadAt.
// The returned value is always the same and is not affected by calls
// to any other method.
func (r *Reader) Size() int64 { return r.size }

// Read implements the io.Reader interface.
func (r *Reader) Read(b []byte) (n int, err error) {
	if r.i >= r.size {
		return 0, io.EOF
	}
	r.prevRune = -1
	blen := len(b)
	for int(r.bi) < len(r.bufs) {
		bi := r.bi
		off := r.offs[bi]
		nn := copy(b[n:], r.bufs[bi][r.i-off:])
		n += nn
		r.advance(int64(nn))
		if n == blen {
			break
		}
	}
	return
}

// ReadAt implements the io.ReaderAt interface.
func (r *Reader) ReadAt(b []byte, off int64) (n int, err error) {
	// cannot modify state - see io.ReaderAt
	if off < 0 {
		return 0, errors.New("bbp.Reader.ReadAt: negative offset")
	}
	if off >= r.size {
		return 0, io.EOF
	}

	blen := len(b)
	ri := off
	bi := 0 // index of bufs
	for _, x := range r.offs[1:] {
		if ri >= x {
			ri -= x
			bi++
			continue
		}
		break
	}
	for bi < len(r.bufs) {
		boff := r.offs[bi]
		nn := copy(b[n:], r.bufs[bi][ri-boff:])
		n += nn
		ri += int64(nn)
		if int(ri-boff) == len(r.bufs[bi]) {
			bi++
		}
		if n == blen {
			break
		}
	}
	if n < blen {
		err = io.EOF
	}
	return
}

// ReadByte implements the io.ByteReader interface.
func (r *Reader) ReadByte() (byte, error) {
	r.prevRune = -1
	if r.i >= r.size {
		return 0, io.EOF
	}
	for int(r.bi) < len(r.bufs) {
		bi := r.bi
		off := r.offs[bi]
		if int(r.i-off) < len(r.bufs[bi]) {
			b := r.bufs[bi][r.i-off]
			r.advance(1)
			return b, nil
		}
		r.bi++
	}
	return 0, io.EOF
}

// UnreadByte complements ReadByte in implementing the io.ByteScanner interface.
func (r *Reader) UnreadByte() error {
	if r.i <= 0 {
		return errors.New("bbp.Reader.UnreadByte: at beginning of slice")
	}
	r.prevRune = -1
	r.moveback(1)
	return nil
}

// ReadRune implements the io.RuneReader interface.
func (r *Reader) ReadRune() (ch rune, size int, err error) {
	if r.i >= r.size {
		r.prevRune = -1
		return 0, 0, io.EOF
	}
	r.prevRune = r.i

	var tmp [4]byte
	r.ReadAt(tmp[:], r.i)
	if tmp[0] < utf8.RuneSelf {
		r.advance(1)
		return rune(tmp[0]), 1, nil
	}
	ch, size = utf8.DecodeRune(tmp[:])
	r.advance(int64(size))
	return
}

// UnreadRune complements ReadRune in implementing the io.RuneScanner interface.
func (r *Reader) UnreadRune() error {
	if r.i <= 0 {
		return errors.New("bbp.Reader.UnreadRune: at beginning of slice")
	}
	if r.prevRune < 0 {
		return errors.New("bbp.Reader.UnreadRune: previous operation was not ReadRune")
	}
	r.advance(r.i - r.prevRune)
	r.prevRune = -1
	return nil
}

// Seek implements the io.Seeker interface.
func (r *Reader) Seek(offset int64, whence int) (int64, error) {
	r.prevRune = -1
	var abs int64
	switch whence {
	case io.SeekStart:
		abs = offset
	case io.SeekCurrent:
		abs = r.i + offset
	case io.SeekEnd:
		abs = r.size + offset
	default:
		return 0, errors.New("bbp.Reader.Seek: invalid whence")
	}
	if abs < 0 {
		return 0, errors.New("bbp.Reader.Seek: negative position")
	}
	r.i = abs
	bi := sort.Search(len(r.offs), func(i int) bool {
		return r.offs[i] >= r.i
	})
	if bi == len(r.offs) || r.offs[bi] > r.i {
		bi--
	}
	r.bi = int64(bi)
	return abs, nil
}

// WriteTo implements the io.WriterTo interface.
func (r *Reader) WriteTo(w io.Writer) (n int64, err error) {
	r.prevRune = -1
	if r.i >= r.size {
		return 0, nil
	}
	for int(r.bi) < len(r.bufs) {
		bi := r.bi
		off := r.offs[bi]
		m, err := w.Write(r.bufs[bi][r.i-off:])
		if m > len(r.bufs[bi][r.i-off:]) {
			panic("bbp.Reader.WriteTo: invalid Write count")
		}
		if err != nil {
			return n, err
		}
		n += int64(m)
		r.advance(int64(m))
	}
	if err == nil && r.i != r.size {
		err = io.ErrShortWrite
	}
	return
}

// NewReader returns a new Reader reading from bufs.
func NewReader(bufs ...[]byte) *Reader {
	r := &Reader{
		bufs:     bufs,
		offs:     make([]int64, len(bufs)),
		prevRune: -1,
	}
	for i, b := range bufs {
		r.size += int64(len(b))
		if i+1 < len(bufs) {
			r.offs[i+1] = r.size
		}
	}
	return r
}
