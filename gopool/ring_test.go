package gopool

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type Tester struct {
	pool *Ring

	taskCount        int64
	timeoutTaskCount int64

	stopped int32
	wg      sync.WaitGroup
}

func (t *Tester) schedule() {
	defer t.wg.Done()
	for atomic.LoadInt32(&t.stopped) == 0 {
		t.pool.Schedule(func() {
			//sleep := rand.Int63n(int64(time.Second))
			//time.Sleep(200*time.Millisecond + time.Duration(sleep))
			time.Sleep(10 * time.Millisecond)
		})
		if x := atomic.AddInt64(&t.taskCount, 1); x%10000 == 0 {
			log.Println("done schedule tasks:", x)
		}
	}
}

func (t *Tester) scheduleTimeout() {
	defer t.wg.Done()
	for atomic.LoadInt32(&t.stopped) == 0 {
		t.pool.ScheduleTimeout(func() {
			//sleep := rand.Int63n(int64(time.Second))
			//time.Sleep(200*time.Millisecond + time.Duration(sleep))
			time.Sleep(10 * time.Millisecond)
		}, 100*time.Millisecond)
		if x := atomic.AddInt64(&t.timeoutTaskCount, 1); x%10000 == 0 {
			log.Println("done scheduleTimeout tasks:", x)
		}
	}
}

func (t *Tester) report() {
	var totalTimeout int
	for range time.NewTicker(time.Second).C {
		sem := atomic.LoadInt32(&t.pool.sem)
		active, busy, pending, timeout := t.pool.Stats()
		totalTimeout += timeout
		scheduled := atomic.LoadInt64(&t.taskCount) + atomic.LoadInt64(&t.timeoutTaskCount)
		log.Printf("stats: sem=%d active=%d busy=%d pending=%d timeout=%d scheduled=%d\n",
			sem, active, busy, pending, totalTimeout, scheduled)
	}
}

func TestRingPool(t *testing.T) {
	tester := &Tester{
		pool: NewRing(1000, 100, 10),
	}

	for i := 0; i < 10; i++ {
		tester.wg.Add(2)
		go tester.schedule()
		go tester.scheduleTimeout()
	}

	go tester.report()

	time.Sleep(5 * time.Second)
	atomic.StoreInt32(&tester.stopped, 1)
	tester.pool.Stop()
	tester.wg.Wait()

	// wait for final report
	time.Sleep(2 * time.Second)
}

func init() {
	go func() {
		http.ListenAndServe("localhost:12345", nil)
	}()
}
