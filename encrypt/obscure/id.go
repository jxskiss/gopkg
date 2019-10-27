package obscure

import (
	"bytes"
	"errors"
	"reflect"
	"unsafe"
)

const (
	idbase      = 59
	idmaxenclen = 12
)

var ErrInvalidInput = errors.New("obscure: invalid input")

type ID int64

func (id ID) String() string {
	idx := int64(id) % idxlen
	idxchar := idxchars[idx]
	seq := idseqchars[idx%seqlen]
	b := id.encode(int64(id), seq)
	b[0] = idxchar
	return b2s(b)
}

func (id *ID) Decode(v string) error {
	idxchar := v[0]
	if idxchar >= 128 {
		return ErrInvalidInput
	}
	idx := idxdec[idxchar]
	if idx == 0 && idxchar != idxchars[0] {
		return ErrInvalidInput
	}
	dec := &idseqdec[idx%seqlen]
	x, err := id.decode(v[1:], dec)
	if err != nil {
		return err
	}
	*id = ID(x)
	return nil
}

func (_ ID) encode(v int64, chars string) []byte {
	var a [idmaxenclen]byte
	i := len(a)
	u := uint64(v)
	b := uint64(idbase)
	for u >= b {
		i--
		q := u / b
		a[i] = chars[uint(u-q*b)]
		u = q
	}
	i--
	a[i] = chars[uint(u)]
	return a[i-1:]
}

func (_ ID) decode(s string, chars *[128]uint8) (int64, error) {
	var n uint64
	var maxVal = uint64(1)<<63 - 1
	for _, c := range s {
		d := uint64(chars[c])
		n *= uint64(idbase)
		n1 := n + d
		if n1 < n || n1 > maxVal {
			return 0, ErrInvalidInput
		}
		n = n1
	}
	return int64(n), nil
}

func (id ID) MarshalJSON() ([]byte, error) {
	if id == 0 {
		return []byte(`"0"`), nil
	}
	s := id.String()
	b := make([]byte, len(s)+2)
	b[0], b[len(b)-1] = '"', '"'
	copy(b[1:], s)
	return b, nil
}

func (id *ID) UnmarshalJSON(b []byte) error {
	if len(b) == 0 || bytes.Equal(b, []byte(`"0"`)) || bytes.Equal(b, []byte("0")) {
		return nil
	}
	if b[0] == '"' {
		b = b[1:]
	}
	if b[len(b)-1] == '"' {
		b = b[:len(b)-1]
	}
	s := b2s(b)
	if err := id.Decode(s); err != nil {
		return err
	}
	return nil
}

func b2s(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func s2b(s string) []byte {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := &reflect.SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}
	return *(*[]byte)(unsafe.Pointer(bh))
}
