package fastrand

import (
	"math/bits"
	"unsafe"
)

// wyrand: https://github.com/wangyi-fudan/wyhash
type wyrand uint64

const (
	wyp0 uint64 = 0xa0761d6478bd642f
	wyp1 uint64 = 0xe7037ed1a0b428db
	wyp2 uint64 = 0x8ebc6af09c88c6e3
	wyp3 uint64 = 0x589965cc75374cc3
	wyp4 uint64 = 0x1d8e4e27c47d124f
)

func _wymix(a, b uint64) uint64 {
	hi, lo := bits.Mul64(a, b)
	return hi ^ lo
}

func (r *wyrand) Uint64() uint64 {
	*r += wyrand(wyp0)
	return _wymix(uint64(*r), uint64(*r^wyrand(wyp1)))
}

func _wyread(seed uint32, p []byte) {
	r := wyrand(seed)
	intp := *(*[]uint64)(unsafe.Pointer(&p))
	var i, end int
	for i, end = 0, len(p)-8; i < end; i += 8 {
		intp[i>>3] = r.Uint64()
	}
	if i < len(p) {
		u64 := r.Uint64()
		for j := range p[i:] {
			p[i+j] = byte(u64 >> (j * 8))
		}
	}
}
