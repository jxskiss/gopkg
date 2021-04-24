package timingwheel

import (
	"errors"
	"time"
	"unsafe"
)

// Timer represents a single time event.
// When the Timer fires, the current time will be send on C,
// unless the Timer was created by AfterFunc.
// A Timer must be created with NewTimer or AfterFunc.
//
// Timer uses timing wheels in underlying, the timing precision is
// roughly ±10 milliseconds; for high precision Timer, you should use
// time.Timer instead of this.
//
// NOTE: we do not provide Reset method as the standard time package
// provides. The doc of time.Timer.Reset says "it is not possible to
// use Reset's return value correctly, as there is a race condition
// between draining the channel and the new timer expiring" and
// "the return value exists to preserve compatibility with existing
// programs". We don't provide Reset method here to keep the timer
// implementation as simple as possible.
type Timer struct {
	C <-chan time.Time
	r *timer
}

// Stop prevents the Timer from firing.
// It returns true if the call stops the timer, false if the timer has
// already expired or been stopped.
// Stop does not close the channel, to prevent a read from the channel
// succeeding incorrectly.
//
// See time.Timer.Stop for more docs abort stopping a timer.
func (t *Timer) Stop() bool {
	return stopTimer(t.r)
}

// NewTimer creates a new Timer that will send the current time on its
// channel after at least duration d.
// The duration d must be greater than zero; if not, NewTimer will
// panic. Stop the ticker to release associated resources.
//
// Timer uses timing wheels in underlying, the timing precision is
// roughly ±10 milliseconds; for high precision Timer, you should use
// time.NewTimer instead of this.
func NewTimer(d time.Duration) *Timer {
	if d <= 0 {
		panic(errors.New("timingwheel: non-positive duration for NewTimer"))
	}
	c := make(chan time.Time, 1)
	r := newTimer(d, 0, t_CH_SEND, unsafe.Pointer(&c))
	addTimer(r)
	t := &Timer{C: c, r: r}
	return t
}

// Ticker holds a channel that delivers ticks of a clock time
// at intervals.
//
// Ticker uses timing wheels in underlying, the timing precision is
// roughly ±10 milliseconds; for high precision Ticker, you should use
// time.Ticker instead of this.
type Ticker struct {
	C <-chan time.Time
	r *timer
}

// Stop turns off a ticker. After Stop, no more ticks will be sent.
// Stop does not close the channel, to prevent a concurrent goroutine
// reading from the channel from seeing an erroneous "tick".
func (t *Ticker) Stop() {
	stopTimer(t.r)
}

// NewTicker returns a new Ticker containing a channel that will send
// the time on the channel after each tick. The period of the ticks is
// specified by the duration argument. The ticker will adjust the time
// interval or drop ticks to make up for slow receivers.
// The duration d must be greater than zero; if not, NewTicker will
// panic. Stop the ticker to release associated resources.
//
// Ticker uses timing wheels in underlying, the timing precision is
// roughly ±10 milliseconds; for high precision Ticker, you should use
// time.NewTicker instead of this.
func NewTicker(d time.Duration) *Ticker {
	if d <= 0 {
		panic(errors.New("timingwheel: non-positive interval for NewTicker"))
	}
	// Give the channel a 1-element time buffer.
	// If the client falls behind while reading, we drop ticks
	// on the floor until the client catches up.
	c := make(chan time.Time, 1)
	r := newTimer(d, d, t_CH_SEND, unsafe.Pointer(&c))
	addTimer(r)
	t := &Ticker{C: c, r: r}
	return t
}

// Sleep pauses the current goroutine for the duration d.
// A negative or zero duration causes Sleep to return immediately.
//
// It uses timing wheels in underlying, the timing precision is roughly
// ±10 milliseconds; for high precision Timer, you should use
// time.Sleep instead of this.
func Sleep(d time.Duration) {
	if d <= 0 {
		return
	}
	<-NewTimer(d).C
}

// After waits for the duration to elapse and then sends the current tme
// on the returned channel.
// It is equivalent to NewTimer(d).C.
//
// The underlying Timer is not recovered by the garbage collector until
// the timer fires. If efficiency is a concern, use NewTimer instead and
// call Timer.Stop is the timer is no longer needed.
//
// The underlying Timer uses timing wheels in underlying, the timing
// precision is roughly ±10 milliseconds; for high precision Timer, you
// should use time.After instead of this.
func After(d time.Duration) <-chan time.Time {
	return NewTimer(d).C
}

// AfterFunc waits for the duration to elapse and then calls f in its own
// goroutine. It returns a Timer that can be used to cancel the call using
// it's Stop method.
//
// The returned Timer uses timing wheels in underlying, the timing precision
// is roughly ±10 milliseconds; for high precision Timer, you should use
// time.AfterFunc instead of this.
func AfterFunc(d time.Duration, f func()) *Timer {
	r := newTimer(d, 0, t_FUNC, unsafe.Pointer(&f))
	addTimer(r)
	t := &Timer{r: r}
	return t
}

// Tick is a convenience wrapper for NewTicker providing access to the ticking
// channel only. While Tick is useful for clients that have no need to shut down
// the Ticker, be aware that without a way to shut it down the underlying
// Ticker cannot be garbage collected; it "leaks".
// Unlike NewTicker, Tick will return nil if d <= 0.
//
// The underlying Ticker uses timing wheels in underlying, the timing
// precision is roughly ±10 milliseconds; for high precision Ticker, you
// should use time.Tick instead of this.
func Tick(d time.Duration) <-chan time.Time {
	if d <= 0 {
		return nil
	}
	return NewTicker(d).C
}

// TickFunc is a convenience wrapper for NewTicker, which calls f in its own
// goroutine for each tick. It returns a Ticker that can be used to stop
// the Ticker using it's Stop method.
//
// The returned Ticker uses timing wheels in underlying, the timing precision
// is roughly ±10 milliseconds; for high precision Ticker, you should use
// time.Ticker functions instead of this.
func TickFunc(d time.Duration, f func()) *Ticker {
	r := newTimer(d, d, t_FUNC, unsafe.Pointer(&f))
	addTimer(r)
	t := &Ticker{r: r}
	return t
}
