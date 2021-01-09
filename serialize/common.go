package serialize

import (
	"encoding/binary"
	"fmt"
	"io"
	"math/bits"
)

var (
	ErrBinaryInvalidFormat  = fmt.Errorf("binary: unexpected bytes format")
	ErrBinaryInvalidLength  = fmt.Errorf("binary: unexpected bytes length")
	ErrProtoInvalidWireType = fmt.Errorf("proto: unexpected wire type")
	ErrProtoInvalidFieldNum = fmt.Errorf("proto: unexpected field num")
	ErrProtoInvalidLength   = fmt.Errorf("proto: invalid negative length")
	ErrProtoIntOverflow     = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEOF        = io.ErrUnexpectedEOF
)

var (
	binEncoding = binary.LittleEndian
	binMagic32  = []byte("EZY0")
	binMagic64  = []byte("EZY1")
)

func encodeVarint(dAtA []byte, offset int, v uint64) int {
	offset -= sov(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}

func sov(x uint64) (n int) {
	return (bits.Len64(x|1) + 6) / 7
}
