package fastrand

import (
	"math/bits"
	"unsafe"
)

// wyrand: https://github.com/wangyi-fudan/wyhash
type wyrand uint64

func _wymix(a, b uint64) uint64 {
	hi, lo := bits.Mul64(a, b)
	return hi ^ lo
}

func (r *wyrand) Uint64() uint64 {
	*r += wyrand(0xa0761d6478bd642f)
	return _wymix(uint64(*r), uint64(*r^wyrand(0xe7037ed1a0b428db)))
}

func _wyread(seed uint32, p []byte) (n int, err error) {
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
	return len(p), nil
}
