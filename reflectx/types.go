package reflectx

import "bytes"

type Bytes []byte

func (p Bytes) String_() string { return b2s(p) }

func (p Bytes) Reader() *bytes.Reader {
	return bytes.NewReader(p)
}

func ToBytes_(s string) Bytes {
	return s2b(s)
}

func String_(b []byte) string {
	return b2s(b)
}
