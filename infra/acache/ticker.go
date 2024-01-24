package acache

import "time"

type callbackTicker interface {
	Stop()
}

type callbackFunc func(time.Time, bool)

func newCallbackTicker(d time.Duration, callback callbackFunc) callbackTicker {
	return newTicker(d, callback)
}

type tickerImpl struct {
	ticker   *time.Ticker
	close    chan struct{}
	callback callbackFunc
}

func newTicker(d time.Duration, callback callbackFunc) *tickerImpl {
	impl := &tickerImpl{
		ticker:   time.NewTicker(d),
		close:    make(chan struct{}),
		callback: callback,
	}
	go impl.run()
	return impl
}

func (t *tickerImpl) Stop() {
	t.ticker.Stop()
	close(t.close)
}

func (t *tickerImpl) run() {
	for {
		select {
		case tick := <-t.ticker.C:
			t.callback(tick, true)
		case <-t.close:
			return
		}
	}
}
