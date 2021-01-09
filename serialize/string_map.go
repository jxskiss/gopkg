package serialize

type StringMap map[string]string

// MarshalProto marshals the string map in format of the following
// protobuf message:
//
// message StringMap {
//     map<string, string> Map = 1;
// }
func (m StringMap) MarshalProto() (dAtA []byte, err error) {
	size := m.ProtoSize()
	dAtA = make([]byte, size)
	n, err := m.MarshalProtoToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m StringMap) MarshalProtoTo(dAtA []byte) (int, error) {
	size := m.ProtoSize()
	return m.MarshalProtoToSizedBuffer(dAtA[:size])
}

func (m StringMap) MarshalProtoToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	if len(m) > 0 {
		for k := range m {
			v := m[k]
			baseI := i
			i -= len(v)
			copy(dAtA[i:], v)
			i = encodeVarint(dAtA, i, uint64(len(v)))
			i--
			dAtA[i] = 0x12
			i -= len(k)
			copy(dAtA[i:], k)
			i = encodeVarint(dAtA, i, uint64(len(k)))
			i--
			dAtA[i] = 0xa
			i = encodeVarint(dAtA, i, uint64(baseI-i))
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func (m StringMap) ProtoSize() (n int) {
	if m == nil {
		return 0
	}
	if len(m) > 0 {
		for k, v := range m {
			mapEntrySize := 1 + len(k) + sov(uint64(len(k))) + 1 + len(v) + sov(uint64(len(v)))
			n += mapEntrySize + 1 + sov(uint64(mapEntrySize))
		}
	}
	return n
}

// UnmarshalProto unmarshalls the string map in format of the following
// protobuf message:
//
// message StringMap {
//     map<string, string> Map = 1;
// }
func (m *StringMap) UnmarshalProto(dAtA []byte) error {
	_map := *m
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
		var msglen int
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrProtoIntOverflow
			}
			if iNdEx >= l {
				return ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			msglen |= int(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		if msglen < 0 {
			return ErrProtoInvalidLength
		}
		postIndex := iNdEx + msglen
		if postIndex < 0 {
			return ErrProtoInvalidLength
		}
		if postIndex > l {
			return ErrUnexpectedEOF
		}
		if _map == nil {
			_map = make(map[string]string)
		}
		var mapkey string
		var mapvalue string
		for iNdEx < postIndex {
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
			if fieldNum == 1 {
				var stringLenmapkey uint64
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrProtoIntOverflow
					}
					if iNdEx >= l {
						return ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					stringLenmapkey |= uint64(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				intStringLenmapkey := int(stringLenmapkey)
				if intStringLenmapkey < 0 {
					return ErrProtoInvalidLength
				}
				postStringIndexmapkey := iNdEx + intStringLenmapkey
				if postStringIndexmapkey < 0 {
					return ErrProtoInvalidLength
				}
				if postStringIndexmapkey > l {
					return ErrUnexpectedEOF
				}
				mapkey = string(dAtA[iNdEx:postStringIndexmapkey])
				iNdEx = postStringIndexmapkey
			} else if fieldNum == 2 {
				var stringLenmapvalue uint64
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrProtoIntOverflow
					}
					if iNdEx >= l {
						return ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					stringLenmapvalue |= uint64(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				intStringLenmapvalue := int(stringLenmapvalue)
				if intStringLenmapvalue < 0 {
					return ErrProtoInvalidLength
				}
				postStringIndexmapvalue := iNdEx + intStringLenmapvalue
				if postStringIndexmapvalue < 0 {
					return ErrProtoInvalidLength
				}
				if postStringIndexmapvalue > l {
					return ErrUnexpectedEOF
				}
				mapvalue = string(dAtA[iNdEx:postStringIndexmapvalue])
				iNdEx = postStringIndexmapvalue
			} else {
				return ErrProtoInvalidFieldNum
			}
		}
		_map[mapkey] = mapvalue
		iNdEx = postIndex
	}

	if iNdEx > l {
		return ErrUnexpectedEOF
	}
	*m = _map
	return nil
}
