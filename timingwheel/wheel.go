package wheel

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

const (
	tvn_bits = 6
	tvr_bits = 8
	tvn_size = 64  // 1 << tvn_bits
	tvr_size = 256 // 1 << tvr_bits

	tvn_mask = 63  // tvn_size - 1
	tvr_mask = 255 // tvr_size -1

	defaultTimerSize = 128
)

type timer struct {
	expires uint64
	period  uint64

	f   func(time.Time, interface{})
	arg interface{}

	w *Wheel

	vec   []*timer
	index int
}

type Wheel struct {
	sync.Mutex

	jiffies uint64

	tv1 [][]*timer
	tv2 [][]*timer
	tv3 [][]*timer
	tv4 [][]*timer
	tv5 [][]*timer

	tick time.Duration

	now   atomic.Value
	tasks chan *tickargs

	quit    chan struct{}
	stopped uint32
}

type tickargs struct {
	now time.Time
	vec []*timer
}

func NewWheel(tick time.Duration) *Wheel {
	f := func(size int) [][]*timer {
		tv := make([][]*timer, size)
		for i := range tv {
			tv[i] = make([]*timer, 0, defaultTimerSize)
		}
		return tv
	}
	w := &Wheel{
		tv1:   f(tvr_size),
		tv2:   f(tvn_size),
		tv3:   f(tvn_size),
		tv4:   f(tvn_size),
		tv5:   f(tvn_size),
		tick:  tick,
		tasks: make(chan *tickargs),
		quit:  make(chan struct{}),
	}
	w.now.Store(time.Now())
	go w.run()
	return w
}

func (w *Wheel) addTimerInternal(t *timer) {
	var expires = t.expires
	var idx = t.expires - w.jiffies
	var tv [][]*timer
	var i uint64

	if idx < tvr_size {
		i = expires & tvr_mask
		tv = w.tv1
	} else if idx < (1 << (tvr_bits + tvn_bits)) {
		i = (expires >> tvr_bits) & tvn_mask
		tv = w.tv2
	} else if idx < (1 << (tvr_bits + 2*tvn_bits)) {
		i = (expires >> (tvr_bits + tvn_bits)) & tvn_mask
		tv = w.tv3
	} else if idx < (1 << (tvr_bits + 3*tvn_bits)) {
		i = (expires >> (tvr_bits + 2*tvn_bits)) & tvn_mask
		tv = w.tv4
	} else if int64(idx) < 0 {
		i = w.jiffies & tvr_mask
		tv = w.tv1
	} else {
		if idx > 0x00000000ffffffff {
			idx = 0x00000000ffffffff
			expires = idx + w.jiffies
		}
		i = (expires >> (tvr_bits + 3*tvn_bits)) & tvn_mask
		tv = w.tv5
	}

	tv[i] = append(tv[i], t)
	t.vec = tv[i]
	t.index = len(tv[i]) - 1
}

func (w *Wheel) cascade(tv [][]*timer, index int) int {
	vec := tv[index]
	tv[index] = vec[0:0:defaultTimerSize]

	for _, t := range vec {
		if t != nil {
			w.addTimerInternal(t)
		}
	}

	return index
}

func (w *Wheel) getIndex(n int) int {
	return int((w.jiffies >> (tvr_bits + uint64(n)*tvn_bits)) & tvn_mask)
}

func (w *Wheel) onTick(now time.Time) {
	w.now.Store(now)

	w.Lock()

	index := int(w.jiffies & tvr_mask)

	if index == 0 &&
		(w.cascade(w.tv2, w.getIndex(0))) == 0 &&
		(w.cascade(w.tv3, w.getIndex(1))) == 0 &&
		(w.cascade(w.tv4, w.getIndex(2))) == 0 &&
		(w.cascade(w.tv5, w.getIndex(3))) == 0 {
	}

	w.jiffies++

	vec := w.tv1[index]
	w.tv1[index] = vec[0:0:defaultTimerSize]

	w.Unlock()

	if len(vec) > 0 {
		args := &tickargs{now, vec}
		select {
		case w.tasks <- args:
		default:
			go w.doTick(args)
		}
	}
}

func (w *Wheel) doTick(args *tickargs) {
	for _, t := range args.vec {
		if t == nil {
			continue
		}
		t.f(args.now, t.arg)
		if t.period > 0 {
			t.expires = t.period + w.jiffies
			w.addTimer(t)
		}
	}
}

func (w *Wheel) addTimer(t *timer) {
	w.Lock()
	w.addTimerInternal(t)
	w.Unlock()
}

func (w *Wheel) delTimer(t *timer) {
	w.Lock()
	vec, idx := t.vec, t.index
	if len(vec) > idx && vec[idx] == t {
		vec[idx] = nil
	}
	w.Unlock()
}

func (w *Wheel) resetTimer(t *timer, when, period time.Duration) {
	w.delTimer(t)

	t.expires = w.jiffies + uint64(when/w.tick)
	t.period = uint64(period / w.tick)

	w.addTimer(t)
}

func (w *Wheel) newTimer(when, period time.Duration, f func(time.Time, interface{}), arg interface{}) *timer {
	t := &timer{
		expires: w.jiffies + uint64(when/w.tick),
		period:  uint64(period / w.tick),
		f:       f,
		arg:     arg,
		w:       w,
	}
	return t
}

func (w *Wheel) run() {
	go func() {
		for args := range w.tasks {
			w.doTick(args)
		}
	}()

	// TODO: is it unnecessary to lock OS thread?
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if w.tick >= time.Second {
		// time.Ticker based loop
		t := time.NewTicker(w.tick)
		defer t.Stop()
		for {
			select {
			case now := <-t.C:
				w.onTick(now)
			case <-w.quit:
				return
			}
		}
	} else {
		// syscall or timerfd based busy loop on unix/linux for lower CPU usage
		//
		// See: https://github.com/golang/go/issues/27707
		var tickUs = w.tick / time.Microsecond
		Utick(uint(tickUs), func() bool {
			if atomic.LoadUint32(&w.stopped) > 0 {
				return true
			}
			w.onTick(time.Now())
			return false
		})
	}
}

func (w *Wheel) Stop() {
	close(w.quit)
}

func (w *Wheel) Sleep(d time.Duration) {
	<-w.NewTimer(d).C
}

func (w *Wheel) After(d time.Duration) <-chan time.Time {
	return w.NewTimer(d).C
}

func (w *Wheel) AfterFunc(d time.Duration, f callback) *Timer {
	t := &Timer{
		r: w.newTimer(d, 0, goFunc, f),
	}
	w.addTimer(t.r)
	return t
}

func (w *Wheel) Tick(d time.Duration) <-chan time.Time {
	return w.NewTicker(d).C
}

func (w *Wheel) TickFunc(d time.Duration, f callback) *Ticker {
	t := &Ticker{
		r: w.newTimer(d, d, goFunc, f),
	}
	w.addTimer(t.r)
	return t
}

func (w *Wheel) NewTimer(d time.Duration) *Timer {
	c := make(chan time.Time, 1)
	t := &Timer{
		C: c,
		r: w.newTimer(d, 0, sendTime, c),
	}
	w.addTimer(t.r)
	return t
}

func (w *Wheel) NewTicker(d time.Duration) *Ticker {
	c := make(chan time.Time, 1)
	t := &Ticker{
		C: c,
		r: w.newTimer(d, d, sendTime, c),
	}
	w.addTimer(t.r)

	return t
}

func sendTime(t time.Time, arg interface{}) {
	select {
	case arg.(chan time.Time) <- t:
	default:
	}
}

type callback func()

func goFunc(t time.Time, arg interface{}) {
	go arg.(callback)()
}

var defaultMilliWheel *Wheel
var defaultWheelOnce sync.Once

func defaultWheel() *Wheel {
	defaultWheelOnce.Do(func() {
		defaultMilliWheel = NewWheel(time.Millisecond)
	})
	return defaultMilliWheel
}
