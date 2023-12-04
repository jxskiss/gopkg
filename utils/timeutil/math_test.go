package timeutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAddJitter(t *testing.T) {
	for _, j := range []float64{0.1, 0.2, 0.3, 0.5, 0.7} {
		for i := 0; i < 100; i++ {
			d := time.Second
			_min := time.Duration(float64(d) * (1 - j))
			_max := time.Duration(float64(d) * (1 + j))
			got := AddJitter(d, j)
			assert.Truef(t, got >= _min && got < _max,
				"j= %v, d= %v, min= %v, max= %v, got= %v", j, d, _min, _max, got)
		}
	}
}

func TestBackoff(t *testing.T) {
	got1, j1 := Backoff(10*time.Second, 0, 0.1)
	assert.Equal(t, 20*time.Second, got1)
	assert.True(t, j1 >= 18*time.Second && j1 < 22*time.Second)

	got2, j2 := Backoff(20*time.Second, 30*time.Second, 0.1)
	assert.Equal(t, 30*time.Second, got2)
	assert.True(t, j2 >= 27*time.Second && j2 < 33*time.Second)
}
