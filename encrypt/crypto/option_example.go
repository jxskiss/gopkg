//go:build ignore

package crypto

import (
	"github.com/jxskiss/base62"
	"golang.org/x/text/encoding/charmap"
)

func ExampleLatinOption() {
	ciphertext, _ := CFBEncrypt(plaintext, key, Encoder(latin1ToUTF8))
	// ...

	plaintext, _ := CFBDecrypt(ciphertext, key, Decoder(utf8ToLatin1))
	// ...
}

func ExampleBase62Option() {
	ciphertext, _ := CFBEncrypt(plaintext, key, Encoder(encodeBase62))
	// ...

	plaintext, _ := CFBEncrypt(ciphertext, key, Decoder(decodeBase62))
	// ...
}

// Be safe with legacy python code by convert arbitrary bytes to utf8.

func latin1ToUTF8(data []byte) ([]byte, error) {
	return charmap.ISO8859_1.NewDecoder().Bytes(data)
}

func utf8ToLatin1(data []byte) ([]byte, error) {
	return charmap.ISO8859_1.NewEncoder().Bytes(data)
}

// Convert arbitrary bytes to Base62 and vice-versa.

func encodeBase62(data []byte) ([]byte, error) {
	out := base62.Encode(data)
	return out, nil
}

func decodeBase62(data []byte) ([]byte, error) {
	return base62.Decode(data)
}
