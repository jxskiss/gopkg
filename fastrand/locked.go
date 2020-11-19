package fastrand

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"unsafe"
)

/*
This file is copied nearly verbatim from https://go-review.googlesource.com/c/go/+/43611/.
See https://github.com/golang/go/issues/20387 for the discussions.
*/

var globalRand = newRand()

func Seed(seed int64)                    { globalRand.Seed(seed) }
func Int63() int64                       { return globalRand.Int63() }
func Uint32() uint32                     { return globalRand.Uint32() }
func Uint64() uint64                     { return globalRand.Uint64() }
func Int31() int32                       { return globalRand.Int31() }
func Int() int                           { return globalRand.Int() }
func Int63n(n int64) int64               { return globalRand.Int63n(n) }
func Int31n(n int32) int32               { return globalRand.Int31n(n) }
func Intn(n int) int                     { return globalRand.Intn(n) }
func Float64() float64                   { return globalRand.Float64() }
func Float32() float32                   { return globalRand.Float32() }
func Perm(n int) []int                   { return globalRand.Perm(n) }
func Shuffle(n int, swap func(i, j int)) { globalRand.Shuffle(n, swap) }
func Read(p []byte) (n int, err error)   { return globalRand.Read(p) }
func NormFloat64() float64               { return globalRand.NormFloat64() }
func ExpFloat64() float64                { return globalRand.ExpFloat64() }

// New returns a new Rand that is safe for concurrent use.
func New(seed int64) *Rand {
	_rnd := newRand()
	_rnd.Seed(seed)
	return _rnd
}

func newRand() *Rand {
	src := newLockedSource()
	return &Rand{src: src, rnd: rand.New(src)}
}

type Rand struct {
	src *lockedSource
	rnd *rand.Rand

	// readVal contains remainder of 63-bit integer used for bytes
	// generation during most recent Read call.
	// It is saved so next Read call can start where the previous
	// one finished.
	readVal int64
	// readPos indicates the number of low-order bytes of readVal
	// that are still valid.
	readPos int8
}

func (r *Rand) Int63() int64                       { return r.rnd.Int63() }
func (r *Rand) Uint32() uint32                     { return r.rnd.Uint32() }
func (r *Rand) Uint64() uint64                     { return r.rnd.Uint64() }
func (r *Rand) Int31() int32                       { return r.rnd.Int31() }
func (r *Rand) Int() int                           { return r.rnd.Int() }
func (r *Rand) Int63n(n int64) int64               { return r.rnd.Int63n(n) }
func (r *Rand) Int31n(n int32) int32               { return r.rnd.Int31n(n) }
func (r *Rand) Intn(n int) int                     { return r.rnd.Intn(n) }
func (r *Rand) Float64() float64                   { return r.rnd.Float64() }
func (r *Rand) Float32() float32                   { return r.rnd.Float32() }
func (r *Rand) Perm(n int) []int                   { return r.rnd.Perm(n) }
func (r *Rand) Shuffle(n int, swap func(i, j int)) { r.rnd.Shuffle(n, swap) }
func (r *Rand) NormFloat64() float64               { return r.rnd.NormFloat64() }
func (r *Rand) ExpFloat64() float64                { return r.rnd.ExpFloat64() }

func (r *Rand) Seed(seed int64) {
	r.src.seedPos(seed, &r.readPos)
}

func (r *Rand) Read(p []byte) (n int, err error) {
	return r.src.read(p, &r.readVal, &r.readPos)
}

func newLockedSource() *lockedSource {
	ls := new(lockedSource)
	for i := range ls.srcs {
		// TODO: What are good initial values?
		ls.srcs[i].Source64 = rand.NewSource(1 + int64(i)*104729).(rand.Source64)
	}
	return ls
}

const nLockedSources = 64

type lockedSource struct {
	n    uint32
	_    [128 - 32]byte
	srcs [nLockedSources]locksource
}

type locksource struct {
	sync.Mutex
	rand.Source64
	_ [128 - unsafe.Sizeof(struct {
		sync.Mutex
		rand.Source64
	}{})]byte
}

func (r *lockedSource) enter() (*locksource, bool) {
	idx := atomic.AddUint32(&r.n, 1) - 1
	return &r.srcs[idx%nLockedSources], idx == 0
}

func (r *lockedSource) exit() {
	// possibly still serial; attempt to detect concurrency.
	// use load-then-store intead of compare-and-swap
	// because it is more performant.
	// If there is a logic race, it will result
	// in unnecessarily setting r.n to zero,
	// i.e. a false positive for being serial, which is ok.
	if atomic.LoadUint32(&r.n) == 1 {
		atomic.StoreUint32(&r.n, 0)
	}
}

func (r *lockedSource) Int63() (n int64) {
	ls, ser := r.enter()
	ls.Lock()
	n = ls.Int63()
	ls.Unlock()
	if ser {
		r.exit()
	}
	return
}

func (r *lockedSource) Uint64() (n uint64) {
	ls, ser := r.enter()
	ls.Lock()
	n = ls.Uint64()
	ls.Unlock()
	if ser {
		r.exit()
	}
	return
}

func (r *lockedSource) Seed(seed int64) {
	atomic.StoreUint32(&r.n, 0)
	ls, ser := r.enter()
	ls.Lock()
	ls.Seed(seed)
	ls.Unlock()
	if ser {
		r.exit()
	}
}

// seedPos implements Seed for a lockedSource without a race condition.
func (r *lockedSource) seedPos(seed int64, readPos *int8) {
	atomic.StoreUint32(&r.n, 0)
	ls := &r.srcs[0]
	ls.Lock()
	ls.Seed(seed)
	*readPos = 0
	ls.Unlock()
}

// read implements Read for a lockedSource without a race condition.
func (r *lockedSource) read(p []byte, readVal *int64, readPos *int8) (n int, err error) {
	ls := &r.srcs[0]
	ls.Lock()
	n, err = read(p, ls.Int63, readVal, readPos)
	ls.Unlock()
	return
}

func read(p []byte, int63 func() int64, readVal *int64, readPos *int8) (n int, err error) {
	pos := *readPos
	val := *readVal
	for n = 0; n < len(p); n++ {
		if pos == 0 {
			val = int63()
			pos = 7
		}
		p[n] = byte(val)
		val >>= 8
		pos--
	}
	*readPos = pos
	*readVal = val
	return
}
