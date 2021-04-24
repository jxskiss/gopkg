package timingwheel

import (
	"sync"
	"sync/atomic"
	"time"
)

const (
	lv1Bits = 8
	lv1Size = 256 // 1 << lv1Bits
	lv1Mask = 255 // lv1Size -1

	lv2Bits = 6
	lv2Size = 64 // 1 << lv2Bits
	lv2Mask = 63 // lv2Size - 1
)

type lv1Timers [lv1Size]*timerlist
type lv2Timers [lv2Size]*timerlist

type wheel struct {
	sync.Mutex

	tick time.Duration

	jiffies uint64

	tv1 lv1Timers
	tv2 lv2Timers
}

type tickTasks struct {
	now    time.Time
	timers *timerlist
}

func (w *wheel) setup() {
	setup := func(tv []*timerlist) {
		for i := len(tv) - 1; i >= 0; i-- {
			tv[i] = &timerlist{w: w}
		}
	}
	setup(w.tv1[:])
	setup(w.tv2[:])
}

func (w *wheel) dispatch(expires uint64) *timerlist {
	var idx = expires - w.jiffies
	var vec *timerlist

	var i uint64
	if idx < lv1Size {
		i = expires & lv1Mask
		vec = w.tv1[i]
		return vec
	}
	i = (expires >> lv1Bits) & lv2Mask
	vec = w.tv2[i]
	return vec
}

func (w *wheel) cascade(tv []*timerlist, index int) int {
	vec := tv[index]
	for i := 0; i < vec.size; {
		t := vec.get2(i)
		next := w.dispatch(t.expires)
		if next == vec {
			i++
			continue
		}

		// vec.del will delete the specific timer from it, and move the
		// last timer to the old place (i), so we don't increase i here,
		// then we can check the moved timer in next iteration.
		vec.del(t)
		next.add(t)
	}

	return index
}

func (w *wheel) onTick(now time.Time) *tickTasks {
	w.Lock()
	defer w.Unlock()

	idx := int(w.jiffies & lv1Mask)
	if idx == 0 {
		lv2Idx := int((w.jiffies >> lv1Bits) & lv2Mask)
		w.cascade(w.tv2[:], lv2Idx)
	}

	w.jiffies++

	if w.tv1[idx].size == 0 {
		return nil
	}

	dueTimers := w.tv1[idx]
	w.tv1[idx] = &timerlist{w: w}
	return &tickTasks{
		now:    now,
		timers: dueTimers,
	}
}

func (w *wheel) addTimer(t *timer) {
	w.Lock()
	defer w.Unlock()

	// The timer may already have been stopped.
	if atomic.LoadUint64(&t.status) != waiting {
		return
	}
	w._addTimer(t)
}

func (w *wheel) _addTimer(t *timer) {

	jiffies := w.jiffies
	n := uint64(t.when / w.tick)
	expires := jiffies + n
	if n > 0 {
		expires -= 1
	}
	t.expires = expires
	vec := w.dispatch(expires)
	vec.add(t)
}

func (w *wheel) delTimer(t *timer) {
	w.Lock()
	w._delTimer(t)
	w.Unlock()
}

func (w *wheel) _delTimer(t *timer) {
	vec, idx := t.vec, t.index
	if vec.get(idx) == t {
		vec.del(t)
	}
}
