package strutil

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jxskiss/gopkg/v2/perf/fastrand"
)

func TestRandomCrypto(t *testing.T) {
	for i := 0; i < 10; i++ {
		length := 3 + fastrand.Intn(1024)
		pwd := RandomCrypto(AlphaDigits, length)
		assert.True(t, IsASCII(pwd))
		assert.Len(t, pwd, length)
	}

	for i := 0; i < 10; i++ {
		length := 3 + fastrand.Intn(1024)
		pwd := RandomCrypto(AlphaDigits+PasswordSpecial, length)
		assert.True(t, IsASCII(pwd))
		assert.Len(t, pwd, length)
	}
}

func TestRandom(t *testing.T) {
	for i := 0; i < 10; i++ {
		length := 3 + fastrand.Intn(1024)
		letters := Random(Letters, length)
		assert.True(t, IsPrintable(letters))
		assert.Len(t, letters, length)
	}

	for i := 0; i < 10; i++ {
		length := 3 + fastrand.Intn(1024)
		lower := Random(LowerLetters, length)
		t.Log("lower:", lower)
		assert.True(t, IsLower(lower))
		assert.Len(t, lower, length)
	}

	for i := 0; i < 10; i++ {
		length := 3 + fastrand.Intn(1024)
		upper := Random(UpperLetters, length)
		assert.True(t, IsUpper(upper))
		assert.Len(t, upper, length)
	}

	for i := 0; i < 10; i++ {
		length := 3 + fastrand.Intn(1024)
		digits := Random(Digits, length)
		assert.NotEqual(t, digits[0], '0')
		assert.True(t, IsASCIIDigit(digits))
		assert.True(t, IsDigit(digits))
		assert.Len(t, digits, length)
	}

	for i := 0; i < 10; i++ {
		length := 3 + fastrand.Intn(1024)
		str := Random(AlphaDigits, length)
		assert.True(t, IsASCII(str))
		assert.True(t, IsPrintable(str))
		assert.Len(t, str, length)
	}
}

func TestRandomHex(t *testing.T) {
	assert.Panics(t, func() {
		_ = RandomHex(0)
	})

	for i := 1; i < 1024; i += 1 {
		got := RandomHex(i)
		assert.Len(t, got, i)
	}
}
