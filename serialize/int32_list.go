package serialize

type Int32List []int32

// MarshalProto marshals the integer array in format of the following
// protobuf message:
//
// message Int32List {
//     repeated int32 values = 1;
// }
func (m Int32List) MarshalProto() (dAtA []byte, err error) {
	size := m.ProtoSize()
	dAtA = make([]byte, size)
	n, err := m.MarshalProtoToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m Int32List) MarshalProtoTo(dAtA []byte) (int, error) {
	size := m.ProtoSize()
	return m.MarshalProtoToSizedBuffer(dAtA[:size])
}

func (m Int32List) MarshalProtoToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	if len(m) > 0 {
		dAtA2 := make([]byte, len(m)*10)
		var j1 int
		for _, num1 := range m {
			num := uint64(num1)
			for num >= 1<<7 {
				dAtA2[j1] = uint8(num&0x7f | 0x80)
				num >>= 7
				j1++
			}
			dAtA2[j1] = uint8(num)
			j1++
		}
		i -= j1
		copy(dAtA[i:], dAtA2[:j1])
		i = encodeVarint(dAtA, i, uint64(j1))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m Int32List) ProtoSize() (n int) {
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
// message Int32List {
//     repeated int32 values = 1;
// }
func (m *Int32List) UnmarshalProto(dAtA []byte) error {
	slice := *m
	l := len(dAtA)
	iNdEx := 0
	var wire uint64
	for shift := uint(0); ; shift += 7 {
		if shift >= 64 {
			return ErrIntegerOverflow
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
		var v int32
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntegerOverflow
			}
			if iNdEx >= l {
				return ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			v |= int32(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		slice = append(slice, v)
	} else if wireType == 2 {
		var packedLen int
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntegerOverflow
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
			return ErrInvalidLength
		}
		postIndex := iNdEx + packedLen
		if postIndex < 0 {
			return ErrInvalidLength
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
			slice = make([]int32, 0, elementCount)
		}
		for iNdEx < postIndex {
			var v int32
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntegerOverflow
				}
				if iNdEx >= l {
					return ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= int32(b&0x7F) << shift
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

func (m Int32List) MarshalBinary() ([]byte, error) {
	bufLen := 1 + 4*len(m)
	out := make([]byte, bufLen)
	out[0] = binMagic32
	buf := out[1:]
	for i, x := range m {
		binEncoding.PutUint32(buf[4*i:4*(i+1)], uint32(x))
	}
	return out, nil
}

func (m *Int32List) UnmarshalBinary(buf []byte) error {
	if len(buf) == 0 {
		return nil
	}
	if len(buf) < 1 || buf[0] != binMagic32 {
		return ErrBinaryInvalidFormat
	}
	buf = buf[1:]
	if len(buf)%4 != 0 {
		return ErrInvalidLength
	}
	slice := *m
	if cap(slice)-len(slice) < len(buf)/4 {
		slice = make([]int32, 0, len(buf)/4)
	}
	for i := 0; i < len(buf); i += 4 {
		x := binEncoding.Uint32(buf[i : i+4])
		slice = append(slice, int32(x))
	}
	*m = slice
	return nil
}
