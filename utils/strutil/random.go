package strutil

import (
	cryptorand "crypto/rand"
	"encoding/hex"
	"math/big"
	"math/bits"
	"unsafe"

	"github.com/jxskiss/gopkg/v2/perf/fastrand"
)

func random(table string, length int) []byte {
	buf := make([]byte, length)
	tabLen := len(table)
	for i := range buf {
		buf[i] = table[fastrand.Intn(tabLen)]
	}
	return buf
}

// Random returns a random string of length consisting of characters
// from table.
// It panics if length <= 0 or len(table) <= 1.
func Random(table string, length int) string {
	if length <= 0 || len(table) <= 1 {
		panic("strutil: invalid argument to Random")
	}
	buf := random(table, length)
	return b2s(buf)
}

// See [crypto/rand.Int] about the implementation details.
func cryptoRandom(table string, length int) []byte {
	ret := make([]byte, 0, length)
	_max := big.NewInt(int64(len(table)))

	// bitLen is the maximum bit length needed to encode a value < max.
	// k is the maximum byte length needed to encode a value < max.
	// b is the number of bits in the most significant byte of max-1.
	bitLen := bits.Len(uint(len(table) - 1))
	k := (bitLen + 7) / 8
	b := uint(bitLen % 8)
	if b == 0 {
		b = 8
	}

	buf := make([]byte, k*(length+10))
	n := new(big.Int)

	for {
		_, err := cryptorand.Read(buf)
		if err != nil {
			panic(err)
		}
		for i := 0; i+k <= len(buf); i += k {
			x := buf[i : i+k]

			// Clear bits in the first byte to increase the probability
			// that the candidate is < max.
			x[0] &= uint8(int(1<<b) - 1)

			n.SetBytes(x)
			if n.Cmp(_max) < 0 {
				ret = append(ret, table[n.Int64()])
				if len(ret) == length {
					return ret
				}
			}
		}
	}
}

// RandomCrypto returns a random string of length consisting of
// characters from table.
// It panics if length <= 0 or len(table) <= 1.
func RandomCrypto(table string, length int) string {
	if length <= 0 || len(table) <= 1 {
		panic("strutil: invalid argument to RandomCrypto")
	}
	buf := cryptoRandom(table, length)
	return b2s(buf)
}

func b2s(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// RandomHex returns a random hex string of length consisting of
// cryptographic-safe random bytes.
func RandomHex(length int) string {
	if length <= 0 {
		panic("strutil: invalid argument to RandomHex")
	}
	n := length/2 + 1
	buf := make([]byte, n)
	_, err := cryptorand.Read(buf)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(buf)[:length]
}
