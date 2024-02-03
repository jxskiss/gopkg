//go:build ignore

package crypto

import (
	"github.com/jxskiss/base62"
	"golang.org/x/text/encoding/charmap"
)

func ExampleLatin1Option() {
	ciphertext, err := CFBEncrypt(plaintext, key, Encoder(latin1ToUTF8))
	// ...

	plaintext, err := CFBDecrypt(ciphertext, key, Decoder(utf8ToLatin1))
	// ...
}

// Be safe with legacy python code by convert arbitrary bytes to utf8.

func latin1ToUTF8(data []byte) ([]byte, error) {
	return charmap.ISO8859_1.NewDecoder().Bytes(data)
}

func utf8ToLatin1(data []byte) ([]byte, error) {
	return charmap.ISO8859_1.NewEncoder().Bytes(data)
}

func ExampleBase62Options() {
	ciphertext, err := CFBEncrypt(plaintext, key, base62Option(base62.StdEncoding))
	// ...

	plaintext, err := CFBDecrypt(ciphertext, key, base62Option(base62.StdEncoding))
	// ...
}

func base62Option(enc *base62.Encoding) Option {
	if enc == nil {
		enc = base62.StdEncoding
	}
	return func(opt *options) {
		opt.encoder = func(src []byte) ([]byte, error) {
			dst := enc.Encode(src)
			return dst, nil
		}
		opt.decoder = func(src []byte) ([]byte, error) {
			return enc.Decode(src)
		}
	}
}
