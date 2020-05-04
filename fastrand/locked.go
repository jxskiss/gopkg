package fastrand

import (
	"math/rand"
	"runtime"
	"sync"
	"time"
)

const cacheLineSize = 64

type lockedSource struct {
	_ [cacheLineSize]byte
	sync.Mutex
	rnd *rand.Rand
}

func (r *lockedSource) Int() int {
	r.Lock()
	x := r.rnd.Int()
	r.Unlock()
	return x
}

func (r *lockedSource) Intn(n int) int {
	r.Lock()
	x := r.rnd.Intn(n)
	r.Unlock()
	return x
}

func (r *lockedSource) Int64() int64 {
	r.Lock()
	x := r.rnd.Int63()
	r.Unlock()
	return x
}

func (r *lockedSource) Int63n(n int64) int64 {
	r.Lock()
	x := r.rnd.Int63n(n)
	r.Unlock()
	return x
}

func (r *lockedSource) Float32() float32 {
	r.Lock()
	x := r.rnd.Float32()
	r.Unlock()
	return x
}

func (r *lockedSource) Float64() float64 {
	r.Lock()
	x := r.rnd.Float64()
	r.Unlock()
	return x
}

func (r *lockedSource) NormFloat64() float64 {
	r.Lock()
	x := r.rnd.NormFloat64()
	r.Unlock()
	return x
}

func (r *lockedSource) ExpFloat64() float64 {
	r.Lock()
	x := r.rnd.ExpFloat64()
	r.Unlock()
	return x
}

func NewRand() Rand {
	src := make([]*lockedSource, maxProcs)
	for i := 0; i < maxProcs; i++ {
		src[i] = &lockedSource{
			rnd: rand.New(rand.NewSource(time.Now().UnixNano())),
		}
	}
	return src
}

type Rand []*lockedSource

func (r Rand) Int() int             { return r[procHint()%maxProcs].Int() }
func (r Rand) Intn(n int) int       { return r[procHint()%maxProcs].Intn(n) }
func (r Rand) Int64() int64         { return r[procHint()%maxProcs].Int64() }
func (r Rand) Int63n(n int64) int64 { return r[procHint()%maxProcs].Int63n(n) }
func (r Rand) Float32() float32     { return r[procHint()%maxProcs].Float32() }
func (r Rand) Float64() float64     { return r[procHint()%maxProcs].Float64() }
func (r Rand) NormFloat64() float64 { return r[procHint()%maxProcs].NormFloat64() }
func (r Rand) ExpFloat64() float64  { return r[procHint()%maxProcs].ExpFloat64() }

var (
	maxProcs    int
	defaultRand Rand
)

func init() {
	maxProcs = runtime.GOMAXPROCS(0)
	defaultRand = NewRand()
}

func Int() int             { return defaultRand.Int() }
func Intn(n int) int       { return defaultRand.Intn(n) }
func Int64() int64         { return defaultRand.Int64() }
func Int63n(n int64) int64 { return defaultRand.Int63n(n) }
func Float32() float32     { return defaultRand.Float32() }
func Float64() float64     { return defaultRand.Float64() }
func NormFloat64() float64 { return defaultRand.NormFloat64() }
func ExpFloat64() float64  { return defaultRand.ExpFloat64() }
