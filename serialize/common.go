package serialize

import (
	"encoding/binary"
	"fmt"
	"io"
	"math/bits"
)

var (
	ErrBinaryInvalidFormat  = fmt.Errorf("serialize: unexpected binary format")
	ErrProtoInvalidWireType = fmt.Errorf("serialize: unexpected proto wire type")
	ErrProtoInvalidFieldNum = fmt.Errorf("serialize: unexpected proto field num")
	ErrInvalidLength        = fmt.Errorf("serialize: invalid length")
	ErrIntegerOverflow      = fmt.Errorf("serialize: integer overflow")
	ErrUnexpectedEOF        = io.ErrUnexpectedEOF
)

const (
	binMagic32        byte = '0'
	binMagic64        byte = '1'
	binDiffCompressed byte = '2' // TODO
)

const maxUint32 = 1<<32 - 1

var binEncoding = binary.LittleEndian

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

func encodeZigZag(buf []byte, offset int, v int64) int {
	zigzag := (v << 1) ^ (v >> 63)
	return encodeVarint(buf, offset, uint64(zigzag))
}
