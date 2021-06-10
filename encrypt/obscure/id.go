package obscure

func (p *Obscure) EncodeID(id int64) []byte {
	buf := make([]byte, 8)
	binEnc.PutUint64(buf, uint64(id))
	return p.EncodeToBytes(buf)
}

func (p *Obscure) DecodeID(src []byte) (int64, error) {
	if len(src) == 0 {
		return 0, ErrInvalidInput
	}
	buf, err := p.DecodeBytes(src)
	if err != nil {
		return 0, err
	}
	if len(buf) != 8 {
		return 0, ErrInvalidInput
	}
	id := int64(binEnc.Uint64(buf))
	return id, nil
}
