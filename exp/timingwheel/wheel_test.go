package timingwheel

import (
	"log"
	"testing"
	"time"
)

func TestTimer(t *testing.T) {
	before := time.Now()
	t1 := NewTimer(500 * time.Millisecond)
	<-t1.C

	after := time.Now()
	log.Println(after.Sub(before))
}

func TestTicker(t *testing.T) {
	wait := make(chan struct{}, 100)
	i := 0
	f := func() {
		log.Println(time.Now())
		i++
		if i >= 10 {
			wait <- struct{}{}
		}
	}

	before := time.Now()
	t1 := TickFunc(1000*time.Millisecond, f)
	<-wait
	t1.Stop()

	after := time.Now()
	log.Println(after.Sub(before))
}
