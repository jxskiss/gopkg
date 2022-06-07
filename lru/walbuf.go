package lru

import "sync"

const (
	// walBufSize must be power of two
	walBufSize = 1024
	walSetSize = walBufSize * 2
	walSetMask = walSetSize - 1
)

var walbufpool sync.Pool

func newWalBuf() *walbuf {
	if buf := walbufpool.Get(); buf != nil {
		return buf.(*walbuf)
	}
	return &walbuf{}
}

// 关于 promotion 和 walbuf 的并发安全性
//
// 1. Cache.buf 永远不为 nil, 当 Cache.buf 写满时，promote 方法中创建新的 walbuf
//    并使用 CAS 操作赋值给 Cache.buf, 成功执行 CAS 的 goroutine 负责触发 flushBuf;
// 2. Cache.promote 方法中对 buf.p 原子加一，每个 goroutine 写入自己拿到的索引位置，
//    不同 goroutine 不会同时写入同一个内存位置;
// 3. 当 Cache.promote 方法被调用时，调用者(Get相关方法)一定持有了 RLock,
//    在 flush walbuf 时，会持有排他锁，因此 promote 方法和 flushBuf 方法一定不会
//    同时执行，flushBuf 函数可以排他地读写 walbuf 的数据;
// 4. flushBuf 方法接受的 walbuf 参数是从 Cache.buf 中 CAS 出来的，又因为 promote
//    和 flushBuf 方法的互斥性，因此保证了一个 walbuf 被传递给 flushBuf 方法后，
//    不会被其他任何 goroutine 持有，flushBuf 结束后，可以安全放回 walbufpool 重用;

// walbuf helps to reduce lock-contention of read requests from the cache.
type walbuf struct {
	b [walBufSize]uint32
	s [walSetSize]uint32
	p int32
}

func (wbuf *walbuf) reset() {
	wbuf.p = 0
	for i := range wbuf.s { // memclr
		wbuf.s[i] = 0
	}
}

func (wbuf *walbuf) deduplicate() []uint32 {
	// Note that we have already checked wbuf.p > 0.
	ln := wbuf.p
	if ln > walBufSize {
		ln = walBufSize
	}

	set := fastHashset(wbuf.s)
	b, p := wbuf.b[:], ln-1
	for i := ln - 1; i >= 0; i-- {
		idx := b[i]
		if !set.has(idx) {
			set.add(idx)
			b[p] = idx
			p--
		}
	}
	return b[p+1 : ln]
}

type fastHashset [walSetSize]uint32

// intPhi is for scrambling the values
const intPhi = 0x9E3779B9

func phiMix(x int64) int64 {
	h := x * intPhi
	return h ^ (h >> 16)
}

func (s *fastHashset) add(value uint32) {
	value += 1

	// Manually inline function phiMix.
	h := int64(value) * intPhi
	ptr := h ^ (h >> 16)

	for {
		ptr &= walSetMask
		k := s[ptr]
		if k == 0 {
			s[ptr] = value
			return
		}
		if k == value {
			return
		}
		ptr += 1
	}
}

func (s *fastHashset) has(value uint32) bool {
	value += 1

	// Manually inline function phiMix.
	h := int64(value) * intPhi
	ptr := h ^ (h >> 16)

	for {
		ptr &= walSetMask
		k := s[ptr]
		if k == value {
			return true
		}
		if k == 0 {
			return false
		}
		ptr += 1
	}
}
