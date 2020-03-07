package easy

import "sync"

type BytesPool struct {
	BufSize   int
	Threshold int

	pool sync.Pool
}

func (p *BytesPool) Get() []byte {
	if x, ok := p.pool.Get().([]byte); ok {
		return x
	}
	size := p.BufSize
	if size == 0 {
		size = 32 * 1024
	}
	return make([]byte, size)
}

func (p *BytesPool) Put(buf []byte) {
	if p.Threshold > 0 && len(buf) > p.Threshold {
		return
	}
	p.pool.Put(buf)
}
