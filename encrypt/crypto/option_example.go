//go:build ignore

package crypto

import (
	"golang.org/x/text/encoding/charmap"
)

func ExampleLatinOption() {
	ciphertext, _ := CFBEncrypt(plaintext, key, Encoder(latin1ToUTF8))
	// ...

	plaintext, _ := CFBDecrypt(ciphertext, key, Decoder(utf8ToLatin1))
	// ...
}

// Be safe with legacy python code by convert arbitrary bytes to utf8.

func latin1ToUTF8(data []byte) ([]byte, error) {
	return charmap.ISO8859_1.NewDecoder().Bytes(data)
}

func utf8ToLatin1(data []byte) ([]byte, error) {
	return charmap.ISO8859_1.NewEncoder().Bytes(data)
}
