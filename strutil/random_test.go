package strutil

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandomCrypto(t *testing.T) {
	for i := 0; i < 10; i++ {
		length := 3 + rand.Intn(1024)
		pwd := RandomCrypto(AlphaDigits, length)
		assert.True(t, IsASCII(pwd))
	}

	for i := 0; i < 10; i++ {
		length := 3 + rand.Intn(1024)
		pwd := RandomCrypto(AlphaDigits+PasswordSpecial, length)
		assert.True(t, IsASCII(pwd))
	}
}

func TestRandom(t *testing.T) {
	for i := 0; i < 10; i++ {
		length := 3 + rand.Intn(1024)
		letters := Random(Letters, length)
		assert.True(t, IsPrintable(letters))
	}

	for i := 0; i < 10; i++ {
		length := 3 + rand.Intn(1024)
		lower := Random(LowerLetters, length)
		t.Log("lower:", lower)
		assert.True(t, IsLower(lower))
	}

	for i := 0; i < 10; i++ {
		length := 3 + rand.Intn(1024)
		upper := Random(UpperLetters, length)
		assert.True(t, IsUpper(upper))
	}

	for i := 0; i < 10; i++ {
		length := 3 + rand.Intn(1024)
		digits := Random(Digits, length)
		assert.NotEqual(t, digits[0], '0')
		assert.True(t, IsASCIIDigit(digits))
		assert.True(t, IsDigit(digits))
	}

	for i := 0; i < 10; i++ {
		length := 3 + rand.Intn(1024)
		str := Random(AlphaDigits, length)
		assert.True(t, IsASCII(str))
		assert.True(t, IsPrintable(str))
	}
}
