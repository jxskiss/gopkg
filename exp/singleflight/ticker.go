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

func newCallbackTicker(d time.Duration, callback func()) callbackTicker {
	if d < lowFrequencyThreshold {
		return newStdTicker(d, callback)
	}
	return newManySelectTicker(d, callback)
}

type stdTickerImpl struct {
	ticker   *time.Ticker
	close    chan struct{}
	callback func()
}

func newStdTicker(d time.Duration, callback func()) *stdTickerImpl {
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
		case <-t.ticker.C:
			t.callback()
		case <-t.close:
			return
		}
	}
}

type manySelectTickerImpl struct {
	ticker *time.Ticker
	task   *mselect.Task
}

func newManySelectTicker(d time.Duration, asyncCallback func()) *manySelectTickerImpl {
	initManySelect()
	ticker := time.NewTicker(d)
	task := mselect.NewTask(ticker.C, nil,
		func(_ time.Time, ok bool) {
			if ok {
				asyncCallback()
			}
		})
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
