package lru

import "sync"

const (
	walBufSize    = 512
	fastThreshold = 8
)

var walbufpool sync.Pool

func newWalBuf() *walbuf {
	if buf := walbufpool.Get(); buf != nil {
		buf := buf.(*walbuf)
		buf.p = 0
		return buf
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
	p int32
}

func (wbuf *walbuf) deduplicate() []uint32 {
	// we have already checked wbuf.p > 0
	ln := wbuf.p
	if ln > walBufSize {
		ln = walBufSize
	}

	if ln > fastThreshold {
		return wbuf.deduplicateSlowPath(ln)
	}

	// Fast path? (not benchmarked)
	b, p := wbuf.b[:], ln-2
LOOP:
	for i := ln - 2; i >= 0; i-- {
		idx := b[i]
		for j := ln - 1; j > p; j-- {
			if b[j] == idx {
				continue LOOP
			}
		}
		b[p] = idx
		p--
	}
	return b[p+1 : ln]
}

func (wbuf *walbuf) deduplicateSlowPath(ln int32) []uint32 {
	m := make(map[uint32]struct{}, ln/2)
	b, p := wbuf.b[:], ln-1
	for i := ln - 1; i >= 0; i-- {
		idx := b[i]
		if _, ok := m[idx]; !ok {
			m[idx] = struct{}{}
			b[p] = idx
			p--
		}
	}
	return b[p+1 : ln]
}
