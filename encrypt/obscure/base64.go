package obscure

func B64Encode(src []byte) (dst []byte) {
	if len(src) == 0 {
		return
	}
	idx := calcSeqIdx(src)
	idxchar := idxchars[idx]
	enc := b64encodings[idx%seqlen]
	dst = make([]byte, enc.EncodedLen(len(src)+1))
	dst[0] = idxchar
	enc.Encode(dst[1:], src)
	return dst
}

func B64EncodeToString(src []byte) string {
	dst := B64Encode(src)
	return b2s(dst)
}

func B64Decode(src []byte) (dst []byte, err error) {
	if len(src) == 0 {
		return
	}
	idxchar := src[0]
	if idxchar >= 128 {
		return nil, ErrInvalidInput
	}
	idx := idxdec[idxchar]
	if idx == 0 && idxchar != idxchars[0] {
		return nil, ErrInvalidInput
	}
	enc := b64encodings[idx%seqlen]
	dst = make([]byte, enc.DecodedLen(len(src)-1))
	n, err := enc.Decode(dst, src[1:])
	if err != nil {
		return nil, ErrInvalidInput
	}
	return dst[:n], nil
}

func B64DecodeString(src string) (dst []byte, err error) {
	b := s2b(src)
	return B64Decode(b)
}
