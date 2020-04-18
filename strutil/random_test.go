package strutil

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)

func TestRandPassword(t *testing.T) {
	for i := 0; i < 10; i++ {
		length := 3 + rand.Intn(1024)
		pwd := RandPassword(length)
		assert.True(t, IsASCII(pwd))
	}

	for i := 0; i < 10; i++ {
		length := 3 + rand.Intn(1024)
		pwd := RandStrongPassword(length)
		assert.True(t, IsASCII(pwd))
	}
}

func TestRandASCII(t *testing.T) {
	for i := 0; i < 10; i++ {
		length := 3 + rand.Intn(1024)
		letters := RandLetters(length)
		assert.True(t, IsPrintable(letters))
	}

	for i := 0; i < 10; i++ {
		length := 3 + rand.Intn(1024)
		lower := RandAlphaLower(length)
		t.Log("lower:", lower)
		assert.True(t, IsLower(lower))
	}

	for i := 0; i < 10; i++ {
		length := 3 + rand.Intn(1024)
		upper := RandAlphaUpper(length)
		assert.True(t, IsUpper(upper))
	}

	for i := 0; i < 10; i++ {
		length := 3 + rand.Intn(1024)
		digits := RandDigits(length)
		assert.NotEqual(t, digits[0], '0')
		assert.True(t, IsASCIIDigit(digits))
		assert.True(t, IsDigit(digits))
	}

	for i := 0; i < 10; i++ {
		length := 3 + rand.Intn(1024)
		str := RandAlphaDigits(length)
		assert.True(t, IsASCII(str))
		assert.True(t, IsPrintable(str))
	}
}
