package timingwheel

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jxskiss/gopkg/exp/fastrand"
)

const (
	milli10Tick = 10 * time.Millisecond
	secondTick  = 2560 * time.Millisecond
)

var (
	/*
	   - tick: 10 milliseconds
	     min: 10 ms
	     max: 256 * 64 * 10 ms = 163.84 s
	     lv1: 10 ms - 2560 ms
	     lv2: 2560 ms - 163.84 s
	     range: 10 ms - 2560 ms
	*/
	milli10Wheel *shardWheel

	/*
		- tick: 2560 milliseconds
		  min: 2.56 s
		  max: 256 * 64 * 2.56 s = 699.0507 m = 11.6508 h
		  lv1: 2.56 s - 10.9227 m
		  lv2: 10.9227 m - 11.6508 h
		  range: 2.56 s - (max uint64)
	*/
	secondWheel *shardWheel
)

var (
	shardSize int // must be power of two
	shardMask int // shardSize - 1
)

func getWheel(d time.Duration) *wheel {
	initWheels()
	idx := int(fastrand.Fastrand()) & shardMask
	if d >= secondTick {
		return secondWheel.shards[idx]
	}
	return milli10Wheel.shards[idx]
}

var (
	runOnce sync.Once
	runWait = make(chan struct{})
)

func initWheels() {
	runOnce.Do(func() {
		shardSize = runtime.GOMAXPROCS(0)
		shardSize = nextPowerOf2(shardSize)
		shardMask = shardSize - 1

		milli10Wheel = newShardWheel(milli10Tick)
		secondWheel = newShardWheel(secondTick)
		go run()
		<-runWait
	})
}

func run() {
	milli10Ch := time.Tick(milli10Wheel.tick)
	secondCh := time.Tick(secondWheel.tick)
	close(runWait)

	for {
		select {
		case now := <-milli10Ch:
			if atomic.CompareAndSwapUint64(&milli10Wheel.working, 0, 1) {
				go onTick(milli10Wheel, now)
			}
		case now := <-secondCh:
			if atomic.CompareAndSwapUint64(&secondWheel.working, 0, 1) {
				go onTick(secondWheel, now)
			}
		}
	}
}

type shardWheel struct {
	tick   time.Duration
	shards []*wheel

	working uint64
	tasks   []*tickTasks
}

func newShardWheel(tick time.Duration) *shardWheel {
	sw := &shardWheel{
		tick:   tick,
		shards: make([]*wheel, shardSize),
	}
	for i := range sw.shards {
		sw.shards[i] = &wheel{tick: tick}
		sw.shards[i].setup()
	}
	return sw
}

func onTick(sw *shardWheel, now time.Time) {
	defer atomic.StoreUint64(&sw.working, 0)

	tasks := sw.tasks[:0]
	for _, shard := range sw.shards {
		args := shard.onTick(now)
		if args == nil {
			continue
		}
		tasks = append(tasks, args)
	}
	for _, t := range tasks {
		doTasks(t.now, t.timers)
	}
}

func doTasks(now time.Time, timers *timerlist) {
	nowNano := now.UnixNano()
	for i := 0; i < timers.size; i++ {
		t := timers.get2(i)

		// Reschedule the timer if it's not ready to fire.
		t.when = time.Duration(t.deadline - nowNano)
		if t.when > 5e6 {
			addTimer(t)
			continue
		}

		// Check that the timer has not been stopped.
		if !atomic.CompareAndSwapUint64(&t.status, waiting, fired) {
			continue
		}

		tFuncs[t.typ](now, t.arg) // !!!

		if t.period > 0 {
			// Again, make sure the timer has not been stopped.
			if atomic.CompareAndSwapUint64(&t.status, fired, waiting) {
				t.when = t.period
				t.deadline = nowNano + int64(t.when)
				addTimer(t)
			}
		}
	}
	timers.release()
}

func nextPowerOf2(x int) int {
	if x == 0 {
		return 1
	}

	x--
	x |= x >> 1
	x |= x >> 2
	x |= x >> 4
	x |= x >> 8
	x |= x >> 16

	return x + 1
}
