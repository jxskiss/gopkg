package lptime

import (
	"testing"
	"time"
)

func TestGetterFunctions(t *testing.T) {
	now := time.Now()
	if now.IsZero() {
		t.Error("Now() should not return zero time")
	}

	i64Funcs := []func() int64{
		Unix, UnixMilli, UnixMicro, UnixNano,
	}
	for _, f := range i64Funcs {
		if f() <= 0 {
			t.Errorf("time function should not return zero or negative value")
		}
	}
}

func TestSetPrecision(t *testing.T) {
	// Restore to the default precision after the test
	defer SetPrecision(defaultPrecision)

	testCases := []struct {
		name      string
		precision time.Duration
		wantPanic bool
	}{
		{"100ms", 100 * time.Millisecond, false},
		{"1s", time.Second, false},
		{"zero", 0, true},
		{"negative", -time.Second, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("SetPrecision(%v) should panic", tc.precision)
					}
				}()
			}
			SetPrecision(tc.precision)
		})
	}
}

func TestSetPrecisionBelowMinimum(t *testing.T) {
	// Restore to the default precision after the test
	defer SetPrecision(defaultPrecision)

	// Setting precision below 10ms should not panic
	SetPrecision(1 * time.Millisecond)
	SetPrecision(5 * time.Millisecond)

	// Now() should still work
	now := Now()
	if now.IsZero() {
		t.Error("Now() should work after setting low precision")
	}
}
