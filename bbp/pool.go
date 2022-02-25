package bbp

import (
	"math"
	"sync/atomic"
	"time"
)

const (
	defaultPoolIdx           = minPoolIdx // 64 bytes
	defaultCalibrateCalls    = 1000
	defaultCalibrateInterval = time.Minute
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
	r Recorder
}

// NewPool creates a new Pool instance using given params.
//
// In most cases, declaring a Pool variable is sufficient to initialize
// a Pool.
func NewPool(defaultSize int, calibrateInterval time.Duration) *Pool {
	r := Recorder{
		DefaultSize:       defaultSize,
		CalibrateInterval: calibrateInterval,
	}
	return &Pool{r}
}

// Get returns a Buffer from the pool with dynamic calibrated default
// capacity. The returned Buffer can be put back to the pool by calling
// Pool.Put(buf) which may be reused later.
func (p *Pool) Get() *Buffer {
	idx := p.r.getPoolIdx()
	buf := new(Buffer)
	buf.buf = sizedPools[idx].Get().([]byte)
	return buf
}

// Put puts back a Buffer to the pool for reusing.
//
// The buf mustn't be touched after returning it to the pool.
// Otherwise, data races will occur.
func (p *Pool) Put(buf *Buffer) {
	p.r.Record(len(buf.buf))
	put(buf.buf)
}

// GetLinkBuffer returns a LinkBuffer from the pool with dynamic calibrated
// default capacity. The returned LinkBuffer can be put back to the pool
// by calling Pool.PutLinkBuffer(buf) which may be reused later.
func (p *Pool) GetLinkBuffer() *LinkBuffer {
	idx := p.r.getPoolIdx()
	buf := &LinkBuffer{
		blockSize: 1 << idx,
		poolIdx:   int(idx),
	}
	return buf
}

// PutLinkBuffer puts back a LinkBuffer to the pool for reusing.
//
// The buf mustn't be touched after returning it to the pool.
// Otherwise, data races will occur.
func (p *Pool) PutLinkBuffer(buf *LinkBuffer) {
	p.r.Record(buf.size)

	// manually inline the PutLinkBuffer function
	poolIdx := buf.poolIdx
	for _, bb := range buf.bufs {
		sizedPools[poolIdx].Put(bb[:0])
	}
}

// Recorder helps to record most frequently used buffer size.
// It calibrates the recorded size data in running, thus it can dynamically
// adjust according to recent workload.
type Recorder struct {

	// DefaultSize optionally configs the initial default size to be used.
	// Default is 64 (in bytes).
	DefaultSize int

	// CalibrateInterval optionally configs the interval to do calibrating.
	// Default is one minute.
	CalibrateInterval time.Duration

	poolIdx uintptr

	calls       [poolSize]int32
	calibrating uintptr
	preNano     int64
	preCalls    int32
}

// Size returns the current most frequently used buffer size.
func (p *Recorder) Size() int {
	idx := atomic.LoadUintptr(&p.poolIdx)
	if idx == 0 {
		idx = defaultPoolIdx
	}
	return 1 << idx
}

// Record records a used buffer size n.
//
// The max recordable size is 32MB, if n is larger than 32MB, it records
// 32MB.
func (p *Recorder) Record(n int) {
	idx := indexGet(n)
	if idx >= poolSize {
		idx = poolSize - 1
	}
	if atomic.AddInt32(&p.calls[idx], -1) < 0 {
		p.calibrate()
	}
}

func (p *Recorder) getPoolIdx() uintptr {
	idx := atomic.LoadUintptr(&p.poolIdx)
	if idx == 0 {
		idx = defaultPoolIdx
	}
	return idx
}

func (p *Recorder) calibrate() {
	if !atomic.CompareAndSwapUintptr(&p.calibrating, 0, 1) {
		return
	}

	nowNano := time.Now().UnixNano()
	nextCalls := int32(defaultCalibrateCalls)
	if p.preCalls > 0 {
		interval := defaultCalibrateInterval
		if p.CalibrateInterval > 0 {
			interval = p.CalibrateInterval
		}
		nextCalls = int32(float64(p.preCalls) * float64(interval) / float64(nowNano-p.preNano))
		if nextCalls < defaultCalibrateCalls {
			nextCalls = defaultCalibrateCalls
		} else if nextCalls > math.MaxInt32 {
			nextCalls = math.MaxInt32
		}
	}
	p.preNano = nowNano
	p.preCalls = nextCalls

	var poolIdx = indexGet(p.DefaultSize)
	var maxCalls int32 = math.MaxInt32
	for i := minPoolIdx; i < poolSize; i++ {
		calls := atomic.SwapInt32(&p.calls[i], nextCalls)
		if calls < maxCalls {
			maxCalls = calls
			poolIdx = i
		}
	}
	atomic.StoreUintptr(&p.poolIdx, uintptr(poolIdx))

	atomic.StoreUintptr(&p.calibrating, 0)
}
