package obscure

func B62Encode(src []byte) (dst []byte) {
	if len(src) == 0 {
		return
	}
	idx := calcSeqIdx(src)
	idxchar := idxchars[idx]
	enc := b62encodings[idx%seqlen]
	dst = enc.Encode(src)
	return append([]byte{idxchar}, dst...)
}

func B62EncodeToString(src []byte) string {
	dst := B62Encode(src)
	return b2s(dst)
}

func B62Decode(src []byte) (dst []byte, err error) {
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
	enc := b62encodings[idx%seqlen]
	dst, err = enc.Decode(src[1:])
	if err != nil {
		return nil, ErrInvalidInput
	}
	return dst, nil
}

func B62DecodeString(src string) (dst []byte, err error) {
	b := s2b(src)
	return B62Decode(b)
}
