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
	binDiffCompressed byte = '2'
)

const maxUint32 = 1<<32 - 1

var binEncoding = binary.LittleEndian

func protoEncodeVarint(dAtA []byte, offset int, v uint64) int {
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

func encodeVarint(buf []byte, x uint64) int {
	for {
		if (int64(x) & ^0x7F) == 0 {
			buf = append(buf, byte(x))
			break
		}
		buf = append(buf, byte(x&0x7F)|0x80)
		x >>= 7
	}
	return len(buf)
}

func decodeVarint(buf []byte) (x uint64, n int, err error) {
	var shift uint
	for ; shift < 64; shift += 7 {
		if n >= len(buf) {
			return 0, 0, ErrUnexpectedEOF
		}
		b := uint64(buf[n])
		n++
		x |= (b & 0x7F) << shift
		if (b & 0x80) == 0 {
			return x, n, nil
		}
	}

	// the number is too large to represent in a 64-bit value
	return 0, 0, ErrIntegerOverflow
}

func sov(x uint64) (n int) {
	return (bits.Len64(x|1) + 6) / 7
}

func encodeZigZag(buf []byte, v int64) int {
	zigzag := (v << 1) ^ (v >> 63)
	return encodeVarint(buf, uint64(zigzag))
}

func decodeZigZag(buf []byte) (x int64, n int, err error) {
	u64, n, err := decodeVarint(buf)
	if err == nil {
		x = int64(u64>>1) ^ -(int64(u64) & 1)
	}
	return
}
