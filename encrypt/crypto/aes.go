package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"io"
)

const (
	gcmTagSize           = 16 // crypto/cipher.gcmTagSize
	gcmStandardNonceSize = 12 // crypto/cipher.gcmStandardNonceSize
)

// GCM：伽罗瓦计数器模式
// GCM模式是CTR和GHASH的组合，GHASH操作定义为密文结果与密钥以及消息长度在GF（2^128）域上相乘。
// GCM比CCM的优势是在于更高并行度及更好的性能。
// TLS1.2标准使用的就是AES-GCM算法，并且Intel CPU提供了GHASH的硬件加速功能。
func GCMEncrypt(plainText, key []byte, opts ...Option) (cipherText []byte, err error) {
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

	cipherText = gcm.Seal(nil, nonce, plainText, opt.additionalData)
	cipherText = append(nonce, cipherText...)
	cipherText, err = opt.encode(cipherText)
	if err != nil {
		return nil, err
	}
	return cipherText, nil
}

func GCMEncryptNewKey(plainText []byte, opts ...Option) (
	cipherText, key, additional []byte, err error,
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

	cipherText = gcm.Seal(nil, nonce, plainText, additional)
	cipherText = append(nonce, cipherText...)
	cipherText, err = opt.encode(cipherText)
	if err != nil {
		return nil, nil, nil, err
	}
	return cipherText, key, additional, nil
}

func UnpackGCMCipherText(cipherText []byte, opts ...Option) (cipherData, nonce, tag []byte) {
	opt := (&options{}).apply(opts...)
	nonceSize := defaultInt(opt.nonceSize, gcmStandardNonceSize)
	tagOffset := len(cipherText) - gcmTagSize
	nonce = cipherText[:nonceSize:nonceSize]
	cipherData = cipherText[nonceSize:tagOffset:tagOffset]
	tag = cipherText[tagOffset:]
	return
}

func GCMDecrypt(cipherText, key []byte, opts ...Option) (plainText []byte, err error) {
	opt := (&options{}).apply(opts...)
	key = KeyPadding(key)
	nonceSize := defaultInt(opt.nonceSize, gcmStandardNonceSize)
	cipherText, err = opt.decode(cipherText)
	if err != nil {
		return nil, err
	}
	nonce := cipherText[:nonceSize]
	cipherText = cipherText[nonceSize:]
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCMWithNonceSize(block, nonceSize)
	if err != nil {
		return nil, err
	}
	return gcm.Open(nil, nonce, cipherText, opt.additionalData)
}

// CBC：密码分组链接模式，明文数据需要按分组大小对齐
func CBCEncrypt(plainText, key []byte, opts ...Option) (cipherText []byte, err error) {
	opt := (&options{}).apply(opts...)
	key = KeyPadding(key)
	plainText = PKCS5Padding(plainText, len(key))

	buf := make([]byte, len(key)+len(plainText))
	nonce := buf[:len(key)]
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	encrypter := cipher.NewCBCEncrypter(block, nonce)
	encrypter.CryptBlocks(buf[len(nonce):], plainText)
	cipherText, err = opt.encode(buf)
	if err != nil {
		return nil, err
	}
	return cipherText, nil
}

func CBCDecrypt(cipherText, key []byte, opts ...Option) (plainText []byte, err error) {
	opt := (&options{}).apply(opts...)
	key = KeyPadding(key)
	cipherText, err = opt.decode(cipherText)
	if err != nil {
		return nil, err
	}
	nonce := cipherText[:len(key)]
	cipherText = cipherText[len(key):]
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	decrypter := cipher.NewCBCDecrypter(block, nonce)
	plainText = make([]byte, len(cipherText))
	decrypter.CryptBlocks(plainText, cipherText)
	plainText = PKCS5UnPadding(plainText)
	return plainText, nil
}

// CFB：密文反馈模式，明文数据不需要按分组大小对齐
func CFBEncrypt(plainText, key []byte, opts ...Option) ([]byte, error) {
	opt := (&options{}).apply(opts...)
	key = KeyPadding(key)

	buf := make([]byte, len(key)+len(plainText))
	nonce := buf[:len(key)]
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	encrypter := cipher.NewCFBEncrypter(block, nonce)
	encrypter.XORKeyStream(buf[len(nonce):], plainText)
	cipherText, err := opt.encode(buf)
	if err != nil {
		return nil, err
	}
	return cipherText, nil
}

func CFBDecrypt(cipherText, key []byte, opts ...Option) ([]byte, error) {
	opt := (&options{}).apply(opts...)
	key = KeyPadding(key)
	cipherText, err := opt.decode(cipherText)
	if err != nil {
		return nil, err
	}
	nonce := cipherText[:len(key)]
	cipherText = cipherText[len(key):]
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	decrypter := cipher.NewCFBDecrypter(block, nonce)
	plainText := make([]byte, len(cipherText))
	decrypter.XORKeyStream(plainText, cipherText)
	return plainText, nil
}

func KeyPadding(key []byte) []byte {
	length := len(key)
	if length == 32 || length == 24 || length == 16 {
		return key
	}
	hash := md5.Sum(key)
	switch {
	case length > 32:
		return key[:32]
	case length > 24:
		return append(key[:length:length], hash[:32-length]...)
	case length > 16:
		return append(key[:length:length], hash[:24-length]...)
	default:
		return append(key[:length:length], hash[:16-length]...)
	}
}

func PKCS5Padding(plainText []byte, blockSize int) []byte {
	padding := blockSize - len(plainText)%blockSize         // 需要padding的数目
	padText := bytes.Repeat([]byte{byte(padding)}, padding) // 生成填充文本
	return append(plainText, padText...)
}

func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unPadding := int(origData[length-1])
	return origData[:(length - unPadding)]
}

func defaultInt(x, defaultValue int) int {
	if x == 0 {
		return defaultValue
	}
	return x
}
