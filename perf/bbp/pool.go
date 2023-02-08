package bbp

import (
	"math"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

const (
	defaultPoolIdx           = 10 // 1024 bytes
	defaultCalibrateCalls    = 10000
	defaultCalibrateInterval = 3 * time.Minute
	defaultResizePercentile  = 90
)

// Pool is a byte buffer pool which reuses byte slice. It uses dynamic
// calibrating (which is a little atomic operations) to try best to match
// the workload.
//
// Generally, if the size and capacity is known in advance, you may use
// the exported function Get(length, capacity) to get a properly sized
// byte buffer. However, if the buffer size is uncertain in advance, you may
// want to use this Pool. For different workloads, dedicated Pool instances
// are recommended, the dynamic calibrating will help to reduce memory waste.
//
// All Pool instances share the same underlying sized byte slice pools.
// The byte buffers provided by Pool has a minimum limit of 64B and a
// maximum limit of 32MB, byte slice with size not in the range will be
// allocated directly from the operating system, and won't be recycled
// for reuse.
//
// The zero value for Pool is ready to use. A Pool value shall not be
// copied after initialized.
type Pool struct {
	noCopy noCopy //nolint:unused

	r Recorder

	sp sync.Pool // []byte
	bp sync.Pool // *Buffer
}

// NewPool creates a new Pool instance using given Recorder.
//
// In most cases, declaring a Pool variable is sufficient to initialize
// a Pool.
func NewPool(r Recorder) *Pool {
	return &Pool{r: r}
}

// Get returns a byte slice buffer from the pool.
// The returned buffer may be put back to the pool for reusing.
func (p *Pool) Get() []byte {
	v := p.sp.Get()
	if v != nil {
		return v.([]byte)
	}
	idx := p.r.getPoolIdx()
	ptr := sizedPools[idx].Get().(unsafe.Pointer)
	return _toBuf(ptr, 0)
}

// GetBuffer returns a Buffer from the pool with dynamic calibrated
// default capacity.
// The returned Buffer may be put back to the pool for reusing.
func (p *Pool) GetBuffer() *Buffer {
	v := p.bp.Get()
	if v != nil {
		return v.(*Buffer)
	}
	idx := p.r.getPoolIdx()
	ptr := sizedPools[idx].Get().(unsafe.Pointer)
	return &Buffer{buf: _toBuf(ptr, 0)}
}

// Put puts back a byte slice buffer to the pool for reusing.
//
// The buf mustn't be touched after returning it to the pool,
// otherwise data races will occur.
func (p *Pool) Put(buf []byte) {
	p.r.Record(len(buf))
	if cap(buf) <= maxBufSize {
		p.sp.Put(buf[:0])
	}
}

// PutBuffer puts back a Buffer to the pool for reusing.
//
// The buf mustn't be touched after returning it to the pool,
// otherwise, data races will occur.
func (p *Pool) PutBuffer(buf *Buffer) {
	p.r.Record(len(buf.buf))
	if cap(buf.buf) <= maxBufSize {
		buf.Reset()
		p.bp.Put(buf)
	}
}

// Recorder helps to record most frequently used buffer size.
// It calibrates the recorded size data in running, thus it can dynamically
// adjust according to recent workload.
type Recorder struct {

	// DefaultSize optionally configs the initial default size to be used.
	// Default is 1024 bytes.
	DefaultSize int

	// CalibrateInterval optionally configs the interval to do calibrating.
	// Default is 3 minutes.
	CalibrateInterval time.Duration

	// ResizePercentile optionally configs the percentile to reset the
	// default size when doing calibrating, the value should be in range
	// [50, 100). Default is 90.
	ResizePercentile int

	poolIdx uintptr

	calls       [poolSize]int32
	calibrating uintptr
	preNano     int64
	preCalls    int32
}

// Size returns the current most frequently used buffer size.
func (p *Recorder) Size() int {
	return 1 << p.getPoolIdx()
}

// Record records a used buffer size n.
//
// The max recordable size is 32MB, if n is larger than 32MB, it records
// 32MB.
func (p *Recorder) Record(n int) {
	idx := maxPoolIdx
	if n < maxBufSize {
		idx = indexGet(n)
	}
	if atomic.AddInt32(&p.calls[idx], -1) < 0 {
		p.calibrate()
	}
}

func (p *Recorder) getPoolIdx() int {
	idx := int(atomic.LoadUintptr(&p.poolIdx))
	if idx == 0 {
		idx = p.getDefaultPoolIdx()
	}
	return idx
}

func (p *Recorder) getDefaultPoolIdx() int {
	if p.DefaultSize > 0 {
		return indexGet(p.DefaultSize)
	}
	return defaultPoolIdx
}

func (p *Recorder) getCalibrateInterval() time.Duration {
	if p.CalibrateInterval > 0 {
		return p.CalibrateInterval
	}
	return defaultCalibrateInterval
}

func (p *Recorder) getResizePercentile() int {
	if p.ResizePercentile >= 50 && p.ResizePercentile < 100 {
		return p.ResizePercentile
	}
	return defaultResizePercentile
}

func (p *Recorder) calibrate() {
	if !atomic.CompareAndSwapUintptr(&p.calibrating, 0, 1) {
		return
	}

	preNano := p.preNano
	preCalls := p.preCalls

	nowNano := time.Now().UnixNano()
	nextCalls := int32(defaultCalibrateCalls)
	if preCalls > 0 {
		interval := p.getCalibrateInterval()
		next := uint64(float64(p.preCalls) * float64(interval) / float64(nowNano-preNano))
		if next < defaultCalibrateCalls {
			nextCalls = defaultCalibrateCalls
		} else if next > math.MaxInt32 {
			nextCalls = math.MaxInt32
		} else {
			nextCalls = int32(next)
		}
	}
	p.preNano = nowNano
	p.preCalls = nextCalls

	var poolIdx int
	var calls [poolSize]int32
	var callsSum int64
	for i := minPoolIdx; i < poolSize; i++ {
		c := atomic.SwapInt32(&p.calls[i], nextCalls)
		if preCalls > 0 {
			c = preCalls - c
			if c < 0 {
				c = preCalls
			}
			calls[i] = c
			callsSum += int64(c)
		}
	}
	if preCalls > 0 {
		pctVal := int64(float64(callsSum) * float64(p.getResizePercentile()) / 100)
		callsSum = 0
		for i := minPoolIdx; i < poolSize; i++ {
			callsSum += int64(calls[i])
			if callsSum >= pctVal {
				poolIdx = i
				break
			}
		}
	}
	if poolIdx == 0 {
		poolIdx = p.getDefaultPoolIdx()
	}
	atomic.StoreUintptr(&p.poolIdx, uintptr(poolIdx))

	atomic.StoreUintptr(&p.calibrating, 0)
}

// noCopy may be added to structs which must not be copied
// after the first use.
//
// See https://golang.org/issues/8005#issuecomment-190753527
// for details.
//
// Note that it must not be embedded, due to the Lock and Unlock methods.
type noCopy struct{} //nolint:all

// Lock is a no-op used by -copylocks checker from `go vet`.
func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}
