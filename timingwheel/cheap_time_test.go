package wheel

import (
	"log"
	"testing"
	"time"
)

func TestCheapTime(t *testing.T) {
	log.Println(cheapTime())
	for i := 0; i < 1000; i++ {
		Sleep(time.Millisecond)
	}
	log.Println(cheapTime())
}
