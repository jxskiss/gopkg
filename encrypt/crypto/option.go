package crypto

import (
	"encoding/base32"
	"encoding/base64"

	"github.com/jxskiss/base62"
)

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

// Decoder optionally specifies a decoder function to decode the encrypted
// ciphertext, it returns an Option.
//
// The decoder function should transform bytes returned by the corresponding
// encoder function to its original bytes.
func Decoder(f func([]byte) ([]byte, error)) Option {
	return func(opt *options) { opt.decoder = f }
}

// Base64 specifies the encoder and decoder to use the provided base64
// encoding, it returns an Option.
//
// If enc is nil, it uses base64.StdEncoding.
func Base64(enc *base64.Encoding) Option {
	if enc == nil {
		enc = base64.StdEncoding
	}
	return func(opt *options) {
		opt.encoder = func(src []byte) ([]byte, error) {
			dst := make([]byte, enc.EncodedLen(len(src)))
			enc.Encode(dst, src)
			return dst, nil
		}
		opt.decoder = func(src []byte) ([]byte, error) {
			dst := make([]byte, enc.DecodedLen(len(src)))
			n, err := enc.Decode(dst, src)
			if err != nil {
				return nil, err
			}
			return dst[:n], nil
		}
	}
}

// Base32 specifies the encoder and decoder to use the provided base32
// encoding, it returns an Option.
//
// If enc is nil, it uses base32.StdEncoding.
func Base32(enc *base32.Encoding) Option {
	if enc == nil {
		enc = base32.StdEncoding
	}
	return func(opt *options) {
		opt.encoder = func(src []byte) ([]byte, error) {
			dst := make([]byte, enc.EncodedLen(len(src)))
			enc.Encode(dst, src)
			return dst, nil
		}
		opt.decoder = func(src []byte) ([]byte, error) {
			dst := make([]byte, enc.DecodedLen(len(src)))
			n, err := enc.Decode(dst, src)
			if err != nil {
				return nil, err
			}
			return dst[:n], nil
		}
	}
}

// Base62 specifies the encoder and decoder to use the provided base62
// encoding, it returns an Option.
//
// If enc is nil, it uses base62.StdEncoding.
func Base62(enc *base62.Encoding) Option {
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
