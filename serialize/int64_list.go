package serialize

type Int64List []int64

// MarshalProto marshals the integer array in format of the following
// protobuf message:
//
// message Int64List {
//     repeated int64 values = 1;
// }
func (m Int64List) MarshalProto() (dAtA []byte, err error) {
	size := m.ProtoSize()
	dAtA = make([]byte, size)
	n, err := m.MarshalProtoToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m Int64List) MarshalProtoTo(dAtA []byte) (int, error) {
	size := m.ProtoSize()
	return m.MarshalProtoToSizedBuffer(dAtA[:size])
}

func (m Int64List) MarshalProtoToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	if len(m) > 0 {
		dAtA4 := make([]byte, len(m)*10)
		var j3 int
		for _, num1 := range m {
			num := uint64(num1)
			for num >= 1<<7 {
				dAtA4[j3] = uint8(num&0x7f | 0x80)
				num >>= 7
				j3++
			}
			dAtA4[j3] = uint8(num)
			j3++
		}
		i -= j3
		copy(dAtA[i:], dAtA4[:j3])
		i = encodeVarint(dAtA, i, uint64(j3))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m Int64List) ProtoSize() (n int) {
	if m == nil {
		return 0
	}
	var l int
	if len(m) > 0 {
		for _, e := range m {
			l += sov(uint64(e))
		}
		n += 1 + sov(uint64(l)) + l
	}
	return n
}

// UnmarshalProto unmarshalls the integer array in format of the following
// protobuf message:
//
// message Int64List {
//     repeated int64 values = 1;
// }
func (m *Int64List) UnmarshalProto(dAtA []byte) error {
	slice := *m
	l := len(dAtA)
	iNdEx := 0
	var wire uint64
	for shift := uint(0); ; shift += 7 {
		if shift >= 64 {
			return ErrProtoIntOverflow
		}
		if iNdEx >= l {
			return ErrUnexpectedEOF
		}
		b := dAtA[iNdEx]
		iNdEx++
		wire |= uint64(b&0x7F) << shift
		if b < 0x80 {
			break
		}
	}
	fieldNum := int32(wire >> 3)
	wireType := int(wire & 0x7)
	if fieldNum != 1 {
		return ErrProtoInvalidFieldNum
	}
	if wireType == 0 {
		var v int64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrProtoIntOverflow
			}
			if iNdEx >= l {
				return ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			v |= int64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		slice = append(slice, v)
	} else if wireType == 2 {
		var packedLen int
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrProtoIntOverflow
			}
			if iNdEx >= l {
				return ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			packedLen |= int(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		if packedLen < 0 {
			return ErrProtoInvalidLength
		}
		postIndex := iNdEx + packedLen
		if postIndex < 0 {
			return ErrProtoInvalidLength
		}
		if postIndex > l {
			return ErrUnexpectedEOF
		}
		var elementCount int
		var count int
		for _, integer := range dAtA[iNdEx:postIndex] {
			if integer < 128 {
				count++
			}
		}
		elementCount = count
		if elementCount != 0 && len(slice) == 0 {
			slice = make([]int64, 0, elementCount)
		}
		for iNdEx < postIndex {
			var v int64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrProtoIntOverflow
				}
				if iNdEx >= l {
					return ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			slice = append(slice, v)
		}
	} else {
		return ErrProtoInvalidWireType
	}

	if iNdEx > l {
		return ErrUnexpectedEOF
	}
	*m = slice
	return nil
}

func (m Int64List) MarshalBinary() ([]byte, error) {
	bigint := false
	for _, x := range m {
		if uint64(x) > maxUint32 {
			bigint = true
			break
		}
	}
	if bigint {
		return m.marshalBinary64()
	} else {
		return m.marshalBinary32()
	}
}

func (m Int64List) marshalBinary32() ([]byte, error) {
	bufLen := 1 + 4*len(m)
	out := make([]byte, bufLen)
	out[0] = binMagic32
	i := 1
	for _, x := range m {
		binEncoding.PutUint32(out[i:i+4], uint32(x))
		i += 4
	}
	return out, nil
}

func (m Int64List) marshalBinary64() ([]byte, error) {
	bufLen := 1 + 8*len(m)
	out := make([]byte, bufLen)
	out[0] = binMagic64
	i := 1
	for _, x := range m {
		binEncoding.PutUint64(out[i:i+8], uint64(x))
		i += 8
	}
	return out, nil
}

func (m *Int64List) UnmarshalBinary(buf []byte) error {
	if len(buf) == 0 {
		return nil
	}
	if len(buf) < 1 {
		return ErrBinaryInvalidFormat
	}
	switch buf[0] {
	case binMagic32:
		return m.unmarshalBinary32(buf[1:])
	case binMagic64:
		return m.unmarshalBinary64(buf[1:])
	}
	return ErrBinaryInvalidFormat
}

func (m *Int64List) unmarshalBinary32(buf []byte) error {
	if len(buf)%4 != 0 {
		return ErrBinaryInvalidLength
	}
	slice := *m
	if cap(slice)-len(slice) < len(buf)/4 {
		slice = make([]int64, 0, len(buf)/4)
	}
	for i := 0; i < len(buf); i += 4 {
		x := binEncoding.Uint32(buf[i : i+4])
		slice = append(slice, int64(x))
	}
	*m = slice
	return nil
}

func (m *Int64List) unmarshalBinary64(buf []byte) error {
	if len(buf)%8 != 0 {
		return ErrBinaryInvalidLength
	}
	slice := *m
	if cap(slice)-len(slice) < len(buf)/8 {
		slice = make([]int64, 0, len(buf)/8)
	}
	for i := 0; i < len(buf); i += 8 {
		x := binEncoding.Uint64(buf[i : i+8])
		slice = append(slice, int64(x))
	}
	*m = slice
	return nil
}
