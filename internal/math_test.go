package internal

import "testing"

func TestNextPowerOfTwo(t *testing.T) {
	data := []struct {
		X    uint
		Want uint
	}{
		{0, 1},
		{1, 1},
		{2, 2},
		{3, 4},
		{4, 4},
		{5, 8},
		{7, 8},
		{8, 8},
		{9, 16},
		{1<<16 - 1, 1 << 16},
		{1 << 16, 1 << 16},
		{1<<16 + 1, 1 << 17},
		{1<<32 - 1, 1 << 32},
		{1 << 32, 1 << 32},
		{1<<63 - 1, 1 << 63},
		{1 << 63, 1 << 63},

		// overflows
		//{1<<63 + 1, 1 << 64},
		//{1<<64 - 1, 1 << 64},
		//{1 << 64, 1 << 64},
	}
	for _, c := range data {
		got := NextPowerOfTwo(c.X)
		if got != c.Want {
			t.Fatalf("x= %v, want %v, got %v", c.X, c.Want, got)
		}
	}
}
