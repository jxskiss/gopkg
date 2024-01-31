package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"io"
)

const (
	gcmTagSize           = 16 // crypto/cipher.gcmTagSize
	gcmStandardNonceSize = 12 // crypto/cipher.gcmStandardNonceSize
)

// GCMEncrypt encrypts plaintext with key using the GCM mode.
// The returned ciphertext contains the nonce, encrypted text and
// the additional data authentication tag. If additional data is not
// provided (as an Option), random data will be generated and used.
//
// GCM模式是CTR和GHASH的组合，GHASH操作定义为密文结果与密钥以及消息长度在GF（2^128）域上相乘。
// GCM比CCM的优势是在于更高并行度及更好的性能。
// TLS1.2标准使用的就是AES-GCM算法，并且Intel CPU提供了GHASH的硬件加速功能。
func GCMEncrypt(plaintext, key []byte, opts ...Option) (ciphertext []byte, err error) {
	opt := (&options{}).apply(opts...)
	key = KeyPadding(key)
	nonceSize := defaultInt(opt.nonceSize, gcmStandardNonceSize)
	nonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCMWithNonceSize(block, nonceSize)
	if err != nil {
		return nil, err
	}

	ciphertext = gcm.Seal(nil, nonce, plaintext, opt.additionalData)
	ciphertext = append(nonce, ciphertext...) //nolint:makezero
	ciphertext, err = opt.encode(ciphertext)
	if err != nil {
		return nil, err
	}
	return ciphertext, nil
}

// GCMEncryptNewKey creates a new key and encrypts plaintext with the
// new key using GCM mode.
// The returned ciphertext contains the nonce, encrypted text and
// the additional data authentication tag. If additional data is not
// provided (as an Option), random data will be generated and used.
func GCMEncryptNewKey(plaintext []byte, opts ...Option) (
	ciphertext, key, additional []byte, err error,
) {
	opt := (&options{}).apply(opts...)
	keySize := defaultInt(opt.keySize, 2*aes.BlockSize)
	nonceSize := defaultInt(opt.nonceSize, gcmStandardNonceSize)
	additional = opt.additionalData
	buflen := keySize + nonceSize
	if len(additional) == 0 {
		buflen += aes.BlockSize
	}
	buf := make([]byte, buflen)
	if _, err = io.ReadFull(rand.Reader, buf); err != nil {
		return nil, nil, nil, err
	}
	nonceEnd := keySize + nonceSize
	key = buf[:keySize:keySize]
	nonce := buf[keySize:nonceEnd:nonceEnd]
	if len(additional) == 0 {
		additional = buf[keySize+nonceSize:]
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, nil, err
	}
	gcm, err := cipher.NewGCMWithNonceSize(block, nonceSize)
	if err != nil {
		return nil, nil, nil, err
	}

	ciphertext = gcm.Seal(nil, nonce, plaintext, additional)
	ciphertext = append(nonce, ciphertext...)
	ciphertext, err = opt.encode(ciphertext)
	if err != nil {
		return nil, nil, nil, err
	}
	return ciphertext, key, additional, nil
}

// UnpackGCMCipherText unpacks cipher text returned by GCMEncrypt and
// GCMEncryptNewKey into encrypted text, nonce and authentication tag.
func UnpackGCMCipherText(ciphertext []byte, opts ...Option) (text, nonce, tag []byte) {
	opt := (&options{}).apply(opts...)
	nonceSize := defaultInt(opt.nonceSize, gcmStandardNonceSize)
	tagOffset := len(ciphertext) - gcmTagSize
	nonce = ciphertext[:nonceSize:nonceSize]
	text = ciphertext[nonceSize:tagOffset:tagOffset]
	tag = ciphertext[tagOffset:]
	return
}

// GCMDecrypt decrypts ciphertext returned by GCMEncrypt and GCMEncryptNewKey
// into plain text.
func GCMDecrypt(ciphertext, key []byte, opts ...Option) (plaintext []byte, err error) {
	opt := (&options{}).apply(opts...)
	key = KeyPadding(key)
	nonceSize := defaultInt(opt.nonceSize, gcmStandardNonceSize)
	ciphertext, err = opt.decode(ciphertext)
	if err != nil {
		return nil, err
	}
	nonce := ciphertext[:nonceSize]
	ciphertext = ciphertext[nonceSize:]
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCMWithNonceSize(block, nonceSize)
	if err != nil {
		return nil, err
	}
	return gcm.Open(nil, nonce, ciphertext, opt.additionalData)
}

// CBCEncrypt encrypts plaintext with key using the CBC mode.
// The given plaintext will be padded following the PKCS#5 standard.
// The returned ciphertext contains the nonce and encrypted data.
//
// CBC - 密码分组链接模式，明文数据需要按分组大小对齐。
func CBCEncrypt(plaintext, key []byte, opts ...Option) (ciphertext []byte, err error) {
	opt := (&options{}).apply(opts...)
	key = KeyPadding(key)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	plaintext = PKCS5Padding(plaintext, blockSize)
	buf := make([]byte, blockSize+len(plaintext))
	nonce := buf[:blockSize]
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	encrypter := cipher.NewCBCEncrypter(block, nonce)
	encrypter.CryptBlocks(buf[len(nonce):], plaintext)
	ciphertext, err = opt.encode(buf)
	if err != nil {
		return nil, err
	}
	return ciphertext, nil
}

// CBCDecrypt decrypts ciphertext returned by CBCEncrypt into plain text.
func CBCDecrypt(ciphertext, key []byte, opts ...Option) (plaintext []byte, err error) {
	opt := (&options{}).apply(opts...)
	key = KeyPadding(key)
	ciphertext, err = opt.decode(ciphertext)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	nonce := ciphertext[:blockSize]
	ciphertext = ciphertext[blockSize:]
	decrypter := cipher.NewCBCDecrypter(block, nonce)
	plaintext = make([]byte, len(ciphertext))
	decrypter.CryptBlocks(plaintext, ciphertext)
	plaintext = PKCS5UnPadding(plaintext)
	return plaintext, nil
}

// CFBEncrypt encrypts plaintext with key using the CFB mode.
// The returned cipher text contains the nonce and encrypted data.
//
// CFB - 密文反馈模式，明文数据不需要按分组大小对齐。
func CFBEncrypt(plaintext, key []byte, opts ...Option) ([]byte, error) {
	opt := (&options{}).apply(opts...)
	key = KeyPadding(key)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	buf := make([]byte, blockSize+len(plaintext))
	nonce := buf[:blockSize]
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	encrypter := cipher.NewCFBEncrypter(block, nonce)
	encrypter.XORKeyStream(buf[len(nonce):], plaintext)
	ciphertext, err := opt.encode(buf)
	if err != nil {
		return nil, err
	}
	return ciphertext, nil
}

// CFBDecrypt decrypts ciphertext returned by CFBEncrypt.
func CFBDecrypt(ciphertext, key []byte, opts ...Option) ([]byte, error) {
	opt := (&options{}).apply(opts...)
	key = KeyPadding(key)
	ciphertext, err := opt.decode(ciphertext)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	nonce := ciphertext[:blockSize]
	ciphertext = ciphertext[blockSize:]
	decrypter := cipher.NewCFBDecrypter(block, nonce)
	plaintext := make([]byte, len(ciphertext))
	decrypter.XORKeyStream(plaintext, ciphertext)
	return plaintext, nil
}

// KeyPadding ensures a key's length is either 32, 24 or 16.
// If key's length is greater than 32, it returns the first 32 bytes of key.
// If key's length is not 32, 24 or 16, it appends additional data to key
// using sha256.Sum(key) to make it satisfies the minimal requirement.
func KeyPadding(key []byte) []byte {
	length := len(key)
	if length == 32 || length == 24 || length == 16 {
		return key
	}
	if length > 32 {
		return key[:32]
	}
	hash := sha256.Sum256(key)
	switch {
	case length > 24:
		return append(key[:length:length], hash[:32-length]...)
	case length > 16:
		return append(key[:length:length], hash[:24-length]...)
	default:
		return append(key[:length:length], hash[:16-length]...)
	}
}

// PKCS5Padding appends padding data to plaintext following the PKCS#5 standard.
func PKCS5Padding(plaintext []byte, blockSize int) []byte {
	padding := blockSize - len(plaintext)%blockSize         // 需要padding的数目
	padText := bytes.Repeat([]byte{byte(padding)}, padding) // 生成填充文本
	return append(plaintext, padText...)
}

// PKCS5UnPadding removes padding data from paddedText following the PKCS#5 standard.
func PKCS5UnPadding(paddedText []byte) []byte {
	length := len(paddedText)
	unPadding := int(paddedText[length-1])
	return paddedText[:(length - unPadding)]
}

func defaultInt(x, defaultValue int) int {
	if x == 0 {
		return defaultValue
	}
	return x
}
