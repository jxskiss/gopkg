package crypto

import (
	"encoding/base64"
)

// B64Encoding is used by the Base64 Option which specifies the encoder
// and decoder function.
// URLEncoding instead of RawURLEncoding is chosen here to be nice with other
// languages, such as python which don't support no-padding in stdlib.
var B64Encoding = base64.URLEncoding

type options struct {
	// encode will be called to transform the encrypted data
	encoder func([]byte) ([]byte, error)

	// decode will be called before data being decrypted
	decoder func([]byte) ([]byte, error)

	// use custom nonce size for AES GCM mode, default: 12
	nonceSize int

	// key size of new key for AES GCM mode, must be 16, 24 or 32, default: 32
	keySize int

	// specify additional data for AES GCM mode, default: nil
	additionalData []byte
}

// Option may be used to customize the encrypt and decrypt functions behavior.
type Option func(opt *options)

func (p *options) apply(opts ...Option) *options {
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func (p *options) encode(data []byte) ([]byte, error) {
	if p.encoder != nil {
		return p.encoder(data)
	}
	return data, nil
}

func (p *options) decode(data []byte) ([]byte, error) {
	if p.decoder != nil {
		return p.decoder(data)
	}
	return data, nil
}

// NonceSize optionally specifies the size of nonce, it returns an Option.
func NonceSize(size int) Option {
	return func(opt *options) { opt.nonceSize = size }
}

// KeySize optionally specifies a key size to use with GCMEncryptNewKey,
// it returns an Option.
func KeySize(size int) Option {
	return func(opt *options) { opt.keySize = size }
}

// AdditionalData optionally specifies the additional data to use with
// GCM mode, it returns an Option.
func AdditionalData(data []byte) Option {
	return func(opt *options) { opt.additionalData = data }
}

// Encoder optionally specifies an encoder function to encode the encrypted
// ciphertext, it returns an Option.
//
// The encoder function may transform arbitrary bytes to a new byte slice
// of some form.
func Encoder(f func([]byte) ([]byte, error)) Option {
	return func(opt *options) { opt.encoder = f }
}

// Decoder optionally specifies an decoder function to decode the encrypted
// ciphertext, it returns an Option.
//
// The decoder function should transform bytes returned by the corresponding
// encoder function to it's original bytes.
func Decoder(f func([]byte) ([]byte, error)) Option {
	return func(opt *options) { opt.decoder = f }
}

// Base64 is an Option which specifies URL encoding as the encoder and
// decoder function.
func Base64(opt *options) {
	opt.encoder = encodeBase64
	opt.decoder = decodeBase64
}

func encodeBase64(data []byte) ([]byte, error) {
	out := make([]byte, B64Encoding.EncodedLen(len(data)))
	B64Encoding.Encode(out, data)
	return out, nil
}

func decodeBase64(data []byte) ([]byte, error) {
	out := make([]byte, B64Encoding.DecodedLen(len(data)))
	n, err := B64Encoding.Decode(out, data)
	if err != nil {
		return nil, err
	}
	return out[:n], nil
}

// "golang.org/x/text/encoding/charmap"
// Be safe with legacy python code by convert arbitrary bytes to utf8.

//func latin1ToUTF8(data []byte) ([]byte, error) {
//	return charmap.ISO8859_1.NewDecoder().Bytes(data)
//}
//
//func utf8ToLatin1(data []byte) ([]byte, error) {
//	return charmap.ISO8859_1.NewEncoder().Bytes(data)
//}

// "github.com/jxskiss/base62"
// Convert arbitrary bytes to Base62 and vice-versa.

//func encodeBase62(data []byte) ([]byte, error) {
//	out := base62.Encode(data)
//	return out, nil
//}
//
//func decodeBase62(data []byte) ([]byte, error) {
//	return base62.Decode(data)
//}
