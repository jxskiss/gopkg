package timingwheel

import (
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

const (
	waiting = iota
	fired
	stopped
)

type timerType uint16

const (
	t_CH_SEND timerType = iota
	t_FUNC
)

var tFuncs = [...]func(time.Time, unsafe.Pointer){
	t_CH_SEND: sendTime,
	t_FUNC:    goFunc,
}

// index consumes 8 bytes.
type index struct {
	i   uint32
	j   uint16
	typ timerType
}

// timer is the internal object to schedule with timing wheel.
// It consumes 64 bytes, which is a cache line size.
type timer struct {
	when   time.Duration
	period time.Duration

	deadline int64  // timestamp in nano seconds
	expires  uint64 // calculate and used by wheel

	arg unsafe.Pointer

	vec *timerlist
	index

	status uint64
}

func newTimer(when, period time.Duration, typ timerType, arg unsafe.Pointer) *timer {
	deadline := time.Now().Add(when).UnixNano()
	t := &timer{
		when:     when,
		period:   period,
		deadline: deadline,
		arg:      arg,
		index: index{
			typ: typ,
		},
	}
	return t
}

func addTimer(t *timer) {
	w := getWheel(t.when)
	w.addTimer(t)
}

func stopTimer(t *timer) bool {
	if atomic.CompareAndSwapUint64(&t.status, waiting, stopped) {
		t.vec.w.delTimer(t)
		return true
	}

	// The timer has expired or already been stopped.
	return false
}

func sendTime(t time.Time, arg unsafe.Pointer) {
	// Non-blocking send of time on arg.
	// Used in NewTimer, it cannot block anyway (buffer).
	// Used in NewTicker, dropping sends on the floor is
	// the desired behavior when the reader gets behind,
	// because the sends are periodic.
	C := *(*chan time.Time)(arg)
	select {
	case C <- t:
	default:
	}
}

func goFunc(_ time.Time, arg unsafe.Pointer) {
	f := *(*func())(arg)
	go f()
}

const (
	timerBucketShift = 3
	timerBucketSize  = 1 << timerBucketShift
	timerBucketMask  = timerBucketSize - 1
)

type timerlist struct {
	w      *wheel
	timers [][]*timer
	size   int
}

func (p *timerlist) packIndex(i uint32, j uint16) int {
	return int(i<<timerBucketShift) | int(j)
}

func (p *timerlist) unpackIndex(idx int) (i uint32, j uint16) {
	i = uint32(idx >> timerBucketShift)
	j = uint16(idx & timerBucketMask)
	return
}

func (p *timerlist) get(idx index) *timer {
	if p.packIndex(idx.i, idx.j) >= p.size {
		return nil
	}
	return p.timers[idx.i][idx.j]
}

func (p *timerlist) get2(idx int) *timer {
	if idx >= p.size {
		return nil
	}
	i, j := p.unpackIndex(idx)
	return p.timers[i][j]
}

func (p *timerlist) add(t *timer) {
	idx := p.size
	i, j := p.unpackIndex(idx)
	if j == 0 {
		if cap(p.timers) > int(i) {
			p.timers = p.timers[:i+1]
		} else {
			newBucket := allocTimerBucket()
			p.timers = append(p.timers, newBucket)
		}
	}
	p.timers[i] = append(p.timers[i], t)
	p.size++

	t.vec = p
	t.i, t.j = i, j
}

func (p *timerlist) del(t *timer) {
	i, j := t.i, t.j
	p.timers[i][j] = nil
	p.size--
	if p.size > 0 {
		p.moveLast(i, j)
	}
}

func (p *timerlist) moveLast(dsti uint32, dstj uint16) {
	srci, srcj := p.unpackIndex(p.size)
	t := p.timers[srci][srcj]
	t.i, t.j = dsti, dstj
	p.timers[dsti][dstj] = t
	p.timers[srci][srcj] = nil
}

func (p *timerlist) release() {
	putTimerBuckets(p.timers)
	p.timers = nil
}

var timerBucketPool sync.Pool

func allocTimerBucket() []*timer {
	b := timerBucketPool.Get()
	if b != nil {
		return b.([]*timer)
	}
	return make([]*timer, 0, timerBucketSize)
}

func putTimerBuckets(timers [][]*timer) {
	timers = timers[:cap(timers)]
	for _, b := range timers {
		for i := 0; i < len(b); i++ {
			b[i] = nil
		}
		timerBucketPool.Put(b[:0])
	}
}
