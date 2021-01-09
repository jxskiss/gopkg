package serialize

type StringList []string

// MarshalProto marshals the string array in format of the following
// protobuf message:
//
// message StringList {
//     repeated string values = 1;
// }
func (m StringList) MarshalProto() (dAtA []byte, err error) {
	size := m.ProtoSize()
	dAtA = make([]byte, size)
	n, err := m.MarshalProtoToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m StringList) MarshalProtoTo(dAtA []byte) (int, error) {
	size := m.ProtoSize()
	return m.MarshalProtoToSizedBuffer(dAtA[:size])
}

func (m StringList) MarshalProtoToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	if len(m) > 0 {
		for iNdEx := len(m) - 1; iNdEx >= 0; iNdEx-- {
			i -= len(m[iNdEx])
			copy(dAtA[i:], m[iNdEx])
			i = encodeVarint(dAtA, i, uint64(len(m[iNdEx])))
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func (m StringList) ProtoSize() (n int) {
	if m == nil {
		return 0
	}
	var l int
	if len(m) > 0 {
		for _, s := range m {
			l = len(s)
			n += 1 + l + sov(uint64(l))
		}
	}
	return n
}

// UnmarshalProto unmarshalls the string array in format of the following
// protobuf message:
//
// message StringList {
//     repeated string values = 1;
// }
func (m *StringList) UnmarshalProto(dAtA []byte) error {
	slice := *m
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
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
		if wireType != 2 {
			return ErrProtoInvalidWireType
		}
		var stringLen uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrProtoIntOverflow
			}
			if iNdEx >= l {
				return ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			stringLen |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		intStringLen := int(stringLen)
		if intStringLen < 0 {
			return ErrProtoInvalidLength
		}
		postIndex := iNdEx + intStringLen
		if postIndex < 0 {
			return ErrProtoInvalidLength
		}
		if postIndex > l {
			return ErrUnexpectedEOF
		}
		slice = append(slice, string(dAtA[iNdEx:postIndex]))
		iNdEx = postIndex
	}

	if iNdEx > l {
		return ErrUnexpectedEOF
	}
	*m = slice
	return nil
}
