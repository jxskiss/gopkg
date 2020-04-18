package strutil

import (
	cryptorand "crypto/rand"
	"math/big"
	"math/rand"
	"time"
	"unsafe"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func random(table string, length int) []byte {
	buf := make([]byte, length)
	max := len(table)
	for i := range buf {
		buf[i] = table[rand.Intn(max)]
	}
	return buf
}

func Random(table string, length int) string {
	buf := random(table, length)
	return b2s(buf)
}

// RandLetters returns a random string containing lowercase and uppercase
// alphabetic characters.
func RandLetters(length int) string {
	return Random(Letters, length)
}

// RandAlphaLower returns a random string containing only lowercase
// alphabetic characters.
func RandAlphaLower(length int) string {
	return Random(AlphaLower, length)
}

// RandAlphaUpper returns a random string containing only uppercase
// alphabetic characters.
func RandAlphaUpper(length int) string {
	return Random(AlphaUpper, length)
}

// RandDigits returns a random string containing only digit numbers, the
// first character is guaranteed to be not zero.
func RandDigits(length int) string {
	out := random(Digits, length)
	if out[0] == '0' {
		out[0] = Digits[rand.Intn(9)+1] // 1-9, no zero
	}
	return b2s(out)
}

// RandAlphaDigits returns a random string containing only characters from
// lowercase, uppercase alphabetic characters or digit numbers.
func RandAlphaDigits(length int) string {
	return Random(AlphaDigits, length)
}

func cryptoRandom(table string, length int) []byte {
	buf := make([]byte, length)
	max := int64(len(table))
	for i := range buf {
		n := big.NewInt(max)
		n, err := cryptorand.Int(cryptorand.Reader, n)
		if err != nil {
			panic(err)
		}
		buf[i] = table[n.Int64()]
	}
	return buf
}

// RandPassword returns a random string containing lowercase and uppercase
// alphabetic characters to be used as a password.
func RandPassword(length int) string {
	buf := cryptoRandom(AlphaDigits, length)
	return b2s(buf)
}

// RandStrongPassword returns a random string containing lowercase and uppercase
// alphabetic characters and punctuation to be used as a password.
func RandStrongPassword(length int) string {
	table := AlphaDigits + PunctNoEscape
	buf := cryptoRandom(table, length)
	return b2s(buf)
}

func b2s(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
