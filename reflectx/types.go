package reflectx

type Bytes []byte

func (p Bytes) String_() string { return b2s(p) }

func ToBytes_(b interface{}) Bytes {
	switch b := b.(type) {
	case Bytes:
		return b
	case string:
		return s2b(b)
	case []byte:
		return b
	}
	panic("ToBytes_: invalid type (must be string/[]byte)")
}

func String_(b []byte) string {
	return b2s(b)
}
