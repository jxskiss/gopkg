package strutil

import (
	cryptorand "crypto/rand"
	"encoding/hex"
	"math/big"
	"unsafe"

	"github.com/jxskiss/gopkg/v2/fastrand"
)

func random(table string, length int) []byte {
	buf := make([]byte, length)
	max := len(table)
	for i := range buf {
		buf[i] = table[fastrand.Intn(max)]
	}
	return buf
}

// Random returns a random string of length consisting of characters
// from table.
func Random(table string, length int) string {
	buf := random(table, length)
	return b2s(buf)
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

// RandomCrypto returns a random string of length consisting of
// characters from table.
func RandomCrypto(table string, length int) string {
	buf := cryptoRandom(table, length)
	return b2s(buf)
}

func b2s(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// RandomHex returns a random hex string of length consisting of
// cryptographic-safe random bytes.
func RandomHex(length int) string {
	if length%2 != 0 {
		panic("invalid argument to RandomHex")
	}
	n := length / 2
	buf := make([]byte, n)
	_, err := cryptorand.Read(buf)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(buf)[:length]
}
