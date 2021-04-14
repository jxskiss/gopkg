package bbp

import (
	"math"
	"sync/atomic"
	"time"
)

const (
	defaultBufSize           = 64
	defaultCalibrateCalls    = 1000
	defaultCalibrateInterval = time.Minute
)

// Pool is a byte buffer pool which reuses byte slice. It uses dynamic
// calibrating (which is a little atomic operations) to try best to match
// the workload.
//
// Generally, if the size and capacity is known in advance, you should use
// the exported function Get(length, capacity) to get a properly sized
// byte buffer. However if the buffer size is uncertain in advance, you may
// want to use this Pool. For different workloads, dedicated Pool instances
// are recommended, the dynamic calibrating will help to reduce memory waste.
//
// All Pool instances share same underlying sized byte slice pools.
//
// The zero value for Pool is ready to use. A Pool value shall not be
// copied after initialized.
type Pool struct {

	// DefaultSize optionally configs the initial default size of
	// byte buffer. The value will be dynamically updated when the
	// Pool is being used. Default is 64 (in bytes).
	DefaultSize int64

	// CalibrateInterval optionally configs the interval to do calibrating.
	// Default is one Minute.
	CalibrateInterval time.Duration

	calls       [poolSize]int32
	calibrating int64
	preNano     int64
	preCalls    int32
}

// Get returns a byte buffer from the pool with dynamic calibrated default
// capacity. The returned byte buffer can be put back to the pool by calling
// Pool.Put(buf) which may be reused later.
func (p *Pool) Get() *Buffer {
	size := int(atomic.LoadInt64(&p.DefaultSize))
	b := bpool.Get().(*Buffer)
	b.B = get(0, size)
	return b
}

// Put puts back a byte buffer to the pool for reusing.
//
// The buf mustn't be touched after returning it to the pool.
// Otherwise data races will occur.
func (p *Pool) Put(buf *Buffer) {
	idx := index(len(buf.B))
	if atomic.AddInt32(&p.calls[idx], -1) < 0 {
		p.calibrate()
	}
	Put(buf)
}

func (p *Pool) calibrate() {
	if !atomic.CompareAndSwapInt64(&p.calibrating, 0, 1) {
		return
	}

	nowNano := time.Now().UnixNano()
	nextCalls := int32(defaultCalibrateCalls)
	if p.preCalls > 0 {
		interval := defaultCalibrateInterval
		if p.CalibrateInterval > 0 {
			interval = p.CalibrateInterval
		}
		nextCalls = int32(float64(p.preCalls) / float64(nowNano-p.preNano) * float64(interval))
		if nextCalls < defaultCalibrateCalls {
			nextCalls = defaultCalibrateCalls
		} else if nextCalls > math.MaxInt32 {
			nextCalls = math.MaxInt32
		}
	}
	p.preNano = nowNano
	p.preCalls = nextCalls

	var defaultSize int64 = defaultBufSize
	var maxCalls int32 = math.MaxInt32
	for i := 0; i < poolSize; i++ {
		calls := atomic.SwapInt32(&p.calls[i], nextCalls)
		if calls != 0 && calls < maxCalls {
			maxCalls = calls
			defaultSize = 1 << i
		}
	}
	atomic.StoreInt64(&p.DefaultSize, defaultSize)

	atomic.StoreInt64(&p.calibrating, 0)
}
