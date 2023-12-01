package singleflight

import (
	"sync"
	"time"

	"github.com/jxskiss/gopkg/v2/perf/mselect"
)

// For low frequency ticker, we use the mselect to reduce the number of goroutines.
const lowFrequencyThreshold = 5 * time.Second

var (
	mselOnce sync.Once
	msel     mselect.ManySelect
)

func initManySelect() {
	mselOnce.Do(func() { msel = mselect.New() })
}

type callbackTicker interface {
	Stop()
}

type callbackFunc func(time.Time, bool)

func newCallbackTicker(d time.Duration, callback callbackFunc) callbackTicker {
	if d < lowFrequencyThreshold {
		return newStdTicker(d, callback)
	}
	return newManySelectTicker(d, callback)
}

type stdTickerImpl struct {
	ticker   *time.Ticker
	close    chan struct{}
	callback callbackFunc
}

func newStdTicker(d time.Duration, callback callbackFunc) *stdTickerImpl {
	impl := &stdTickerImpl{
		ticker:   time.NewTicker(d),
		close:    make(chan struct{}),
		callback: callback,
	}
	go impl.run()
	return impl
}

func (t *stdTickerImpl) Stop() {
	t.ticker.Stop()
	close(t.close)
}

func (t *stdTickerImpl) run() {
	for {
		select {
		case tick := <-t.ticker.C:
			t.callback(tick, true)
		case <-t.close:
			return
		}
	}
}

type manySelectTickerImpl struct {
	ticker *time.Ticker
	task   *mselect.Task
}

func newManySelectTicker(d time.Duration, asyncCallback callbackFunc) *manySelectTickerImpl {
	initManySelect()
	ticker := time.NewTicker(d)
	task := mselect.NewTask(ticker.C, nil, asyncCallback)
	msel.Add(task)
	impl := &manySelectTickerImpl{
		ticker: ticker,
		task:   task,
	}
	return impl
}

func (t *manySelectTickerImpl) Stop() {
	t.ticker.Stop()
	msel.Delete(t.task)
}
