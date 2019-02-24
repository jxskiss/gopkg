package gopool

import (
	"log"
	"sync/atomic"
	"testing"
	"time"
)

type Tester struct {
	pool *Ring
}

func (t *Tester) schedule() {
	var count int64
	for {
		t.pool.Schedule(func() {
			//sleep := rand.Int63n(int64(time.Second))
			//time.Sleep(200*time.Millisecond + time.Duration(sleep))
			time.Sleep(10 * time.Millisecond)
		})
		if x := atomic.AddInt64(&count, 1); x%10000 == 0 {
			log.Println("done schedule tasks:", x)
		}
	}
}

func (t *Tester) scheduleTimeout() {
	var count int64
	for {
		t.pool.ScheduleTimeout(func() {
			//sleep := rand.Int63n(int64(time.Second))
			//time.Sleep(200*time.Millisecond + time.Duration(sleep))
			time.Sleep(10 * time.Millisecond)
		}, 100*time.Millisecond)
		if x := atomic.AddInt64(&count, 1); x%10000 == 0 {
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
		log.Printf("stats: sem=%d active=%d busy=%d pending=%d timeout=%d\n",
			sem, active, busy, pending, totalTimeout)
	}
}

func TestRingPool(t *testing.T) {
	tester := &Tester{
		pool: NewRing(1000, 100, 10),
	}

	for i := 0; i < 10; i++ {
		go tester.schedule()
		go tester.scheduleTimeout()
	}
	go tester.report()

	<-tester.pool.stop
}
